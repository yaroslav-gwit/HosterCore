package cmd

import (
	"errors"
	"log"
	"os"
	"os/exec"
	"strconv"

	"github.com/spf13/cobra"
	"golang.org/x/exp/slices"
)

var (
	vmDiskCmd = &cobra.Command{
		Use:   "disk",
		Short: "VM disk image related commands",
		Long:  `VM disk image related commands: add, expand, etc`,
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}
)

var (
	diskImage     string
	expansionSize int

	vmDiskExpandCmd = &cobra.Command{
		Use:   "expand [vmName]",
		Short: "Expand the VM drive",
		Long:  `Expand the VM drive. Can only be done offline due to data loss possibility.`,
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			err := checkInitFile()
			if err != nil {
				log.Fatal(err.Error())
			}

			err = diskExpandOffline(args[0], diskImage, expansionSize)
			if err != nil {
				log.Fatal(err)
			}
		},
	}
)

var (
	vmDiskAddSize int

	vmDiskAddCmd = &cobra.Command{
		Use:   "add [vmName]",
		Short: "Add a new disk image",
		Long:  `Add a new disk image to this VM dataset. Can only be done offline due to the fact that bhyve can't hot-reload settings.`,
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			err := checkInitFile()
			if err != nil {
				log.Fatal(err.Error())
			}

			err = diskAddOffline(args[0], vmDiskAddSize)
			if err != nil {
				log.Fatal(err)
			}
		},
	}
)

func diskExpandOffline(vmName string, diskImage string, expansionSize int) error {
	allVms := getAllVms()
	if slices.Contains(allVms, vmName) {
		_ = 0
	} else {
		return errors.New("vm was not found")
	}

	vmFolder := getVmFolder(vmName)
	vmConfigVar := vmConfig(vmName)
	if vmLiveCheck(vmName) {
		return errors.New("vm has to be offline, due to the data loss possibility of online expansion")
	}

	if vmConfigVar.ParentHost != GetHostName() {
		return errors.New("this host isn't a parent of this vm, please make sure the vm is not a backup from another host")
	}

	diskLocation := vmFolder + "/" + diskImage
	if !diskImageExists(diskLocation) {
		return errors.New("disk image doesn't exist")
	}

	cmd := exec.Command("truncate", "-s", "+"+strconv.Itoa(expansionSize)+"G", diskLocation)
	err := cmd.Run()
	if err != nil {
		return errors.New("can't expand the drive: " + err.Error())
	}

	return nil
}

func diskAddOffline(vmName string, imageSize int) error {
	allVms := getAllVms()
	if !slices.Contains(allVms, vmName) {
		return errors.New("vm was not found")
	}

	vmFolder := getVmFolder(vmName)
	vmConfigVar := vmConfig(vmName)
	if vmLiveCheck(vmName) {
		return errors.New("vm has to be offline, due to the data loss possibility of online expansion")
	}

	if vmConfigVar.ParentHost != GetHostName() {
		return errors.New("this host isn't a parent of this vm, please make sure the vm is not a backup from another host")
	}

	diskIndex := 1
	diskImage := "disk" + strconv.Itoa(diskIndex) + ".img"
	diskLocation := vmFolder + "/" + diskImage
	for {
		if diskImageExists(diskLocation) {
			diskIndex = diskIndex + 1
			diskImage = "disk" + strconv.Itoa(diskIndex) + ".img"
			diskLocation = vmFolder + "/" + diskImage
		} else {
			break
		}
	}

	stdOut, stdErr := exec.Command("touch", diskLocation).CombinedOutput()
	if stdErr != nil {
		return errors.New("could not create an image file: " + string(stdOut) + "; " + stdErr.Error())
	}
	stdOut, stdErr = exec.Command("truncate", "-s", "+"+strconv.Itoa(imageSize)+"G", diskLocation).CombinedOutput()
	if stdErr != nil {
		return errors.New("could not expand an image file: " + string(stdOut) + "; " + stdErr.Error())
	}
	return nil
}

// Returns true if VM disk image exists. Takes in disk absolute path as a parameter.
func diskImageExists(diskLocation string) bool {
	_, err := os.Stat(diskLocation)
	return !os.IsNotExist(err)
}
