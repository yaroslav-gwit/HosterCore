package cmd

import (
	HosterVmUtils "HosterCore/internal/pkg/hoster/vm/utils"
	"errors"
	"log"
	"os"
	"os/exec"
	"strconv"

	"github.com/spf13/cobra"
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
			checkInitFile()
			err := DiskExpandOffline(args[0], diskImage, expansionSize)
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
			checkInitFile()
			err := diskAddOffline(args[0], vmDiskAddSize)
			if err != nil {
				log.Fatal(err)
			}
		},
	}
)

func DiskExpandOffline(vmName string, diskImage string, expansionSize int) error {
	vm, err := HosterVmUtils.InfoJsonApi(vmName)
	if err != nil {
		return err
	}
	// vmFolder := vm.Simple.MountPoint.Mountpoint + "/" + vm.Name
	vmFolder := vm.Simple.Mountpoint + "/" + vm.Name

	if vm.Running {
		return errors.New("vm has to be offline, due to the data loss possibility of online expansion")
	}
	if vm.Backup {
		return errors.New("this is a backup")
	}

	if len(diskImage) < 1 {
		diskImage = "disk0.img"
	}

	diskLocation := vmFolder + "/" + diskImage
	if !diskImageExists(diskLocation) {
		return errors.New("disk image doesn't exist: " + diskLocation)
	}

	cmd := exec.Command("truncate", "-s", "+"+strconv.Itoa(expansionSize)+"G", diskLocation)
	err = cmd.Run()
	if err != nil {
		return errors.New("can't expand the drive: " + err.Error())
	}

	return nil
}

func diskAddOffline(vmName string, imageSize int) error {
	vm, err := HosterVmUtils.InfoJsonApi(vmName)
	if err != nil {
		return err
	}
	vmFolder := vm.Simple.Mountpoint + "/" + vm.Name

	if vm.Running {
		return errors.New("vm has to be offline, due to the data loss possibility of online expansion")
	}
	if vm.Backup {
		return errors.New("this vm is a backup")
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

	var diskConfig HosterVmUtils.VmDisk
	diskConfig.DiskType = "nvme"
	diskConfig.DiskLocation = "internal"
	diskConfig.DiskImage = diskImage
	diskConfig.Comment = "Additional disk image"
	vm.VmConfig.Disks = append(vm.VmConfig.Disks, diskConfig)

	err = HosterVmUtils.ConfigFileWriter(vm.VmConfig, vmFolder+"/"+HosterVmUtils.VM_CONFIG_NAME)
	if err != nil {
		return err
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
