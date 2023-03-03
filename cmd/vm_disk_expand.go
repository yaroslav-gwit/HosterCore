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
	diskImage     string
	expansionSize int

	vmDistExpandCmd = &cobra.Command{
		Use:   "disk-expand [vmName]",
		Short: "Expand the VM drive",
		Long:  `Expand the VM drive. Can only be done offline due to data loss possibility.`,
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			err := diskExpandOffline(args[0], diskImage, expansionSize)
			if err != nil {
				log.Fatal(err)
			}
		},
	}
)

func diskExpandOffline(vmName string, diskImage string, expansionSize int) error {
	vmFolder := getVmFolder(vmName)
	vmConfigVar := vmConfig(vmName)
	if !slices.Contains(getAllVms(), vmName) {
		return errors.New("vm was not found")
	} else if vmLiveCheck(vmName) {
		return errors.New("vm has to be offline, due to the data loss possibility of online expansion")
	} else if vmConfigVar.ParentHost != GetHostName() {
		return errors.New("this host isn't a parent of this vm, please make sure the vm is not a backup from another host")
	}

	diskLocation := vmFolder + "/" + diskImage
	_, err := os.Stat(diskLocation)
	if os.IsNotExist(err) {
		return errors.New("disk image doesn't exist")
	}

	cmd := exec.Command("truncate", "-s", "+"+strconv.Itoa(expansionSize)+"G", diskLocation)
	err = cmd.Run()
	if err != nil {
		return errors.New("can't expand the drive: " + err.Error())
	}

	return nil
}
