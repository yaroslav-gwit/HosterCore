package cmd

import (
	"HosterCore/pkg/emojlog"
	"errors"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

var (
	vmCiIsoCmd = &cobra.Command{
		Use:   "ci-iso",
		Short: "CloudInit ISO related operations",
		Long:  `CloudInit ISO related operations.`,
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			checkInitFile()
			cmd.Help()
		},
	}
)

var (
	vmCiIsoMountCmd = &cobra.Command{
		Use:   "mount [vmName]",
		Short: "Mount CloudInit ISO",
		Long:  `Mount CloudInit ISO.`,
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			checkInitFile()

			err := mountCiIso(args[0])
			if err != nil {
				emojlog.PrintLogMessage(err.Error(), emojlog.Error)
				os.Exit(1)
			}
		},
	}
)

var (
	vmCiIsoUnmountCmd = &cobra.Command{
		Use:   "unmount [vmName]",
		Short: "Unmount CloudInit ISO",
		Long:  `Unmount CloudInit ISO.`,
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			checkInitFile()
			err := unmountCiIso(args[0])
			if err != nil {
				emojlog.PrintLogMessage(err.Error(), emojlog.Error)
				os.Exit(1)
			}
		},
	}
)

func mountCiIso(vmName string) error {
	vmConfigVar := vmConfig(vmName)
	vmFolder := getVmFolder(vmName)
	if vmConfigVar.Disks[1].DiskImage == "seed.iso" {
		return errors.New("CloudInit ISO has already been mounted")
	}

	vmConfigVar.Disks[1].DiskImage = "seed.iso"
	vmConfigVar.Disks[1].Comment = "CloudInit ISO file"
	err := vmConfigFileWriter(vmConfigVar, vmFolder+"/vm_config.json")
	if err != nil {
		return nil
	}

	if VmLiveCheck(vmName) {
		emojlog.PrintLogMessage("Please don't forget to reboot the VM to apply changes", emojlog.Debug)
	}
	emojlog.PrintLogMessage("CloudInit ISO has been mounted", emojlog.Changed)
	return nil
}

func unmountCiIso(vmName string) error {
	fileContents := []byte("placeholder file for an empty CI ISO file")
	vmFolder := getVmFolder(vmName)
	vmConfigVar := vmConfig(vmName)
	if vmConfigVar.Disks[1].DiskImage == "seed-empty.iso" {
		return errors.New("CloudInit ISO has already been unmounted")
	}
	err := os.WriteFile(vmFolder+"/placeholder", fileContents, 0640)
	if err != nil {
		return nil
	}

	_ = os.Remove(vmFolder + "/seed-empty.iso")
	out, err := exec.Command("genisoimage", "-output", vmFolder+"/seed-empty.iso", "-volid", "cidata", "-joliet", "-rock", vmFolder+"/placeholder").CombinedOutput()
	if err != nil {
		return errors.New("there was a problem generating an ISO: " + string(out) + "; " + err.Error())
	}
	_ = os.Remove(vmFolder + "/placeholder")

	vmConfigVar.Disks[1].DiskImage = "seed-empty.iso"
	vmConfigVar.Disks[1].Comment = "An empty CloudInit ISO file"
	err = vmConfigFileWriter(vmConfigVar, vmFolder+"/vm_config.json")
	if err != nil {
		return nil
	}

	if VmLiveCheck(vmName) {
		emojlog.PrintLogMessage("Please don't forget to reboot the VM to apply changes", emojlog.Debug)
	}
	emojlog.PrintLogMessage("CloudInit ISO has been unmounted", emojlog.Changed)
	return nil
}
