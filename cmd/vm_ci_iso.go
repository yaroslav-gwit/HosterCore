//go:build freebsd
// +build freebsd

package cmd

import (
	"HosterCore/internal/pkg/emojlog"
	HosterVmUtils "HosterCore/internal/pkg/hoster/vm/utils"
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
	vmConf, err := HosterVmUtils.InfoJsonApi(vmName)
	if err != nil {
		return err
	}

	vmFolder := vmConf.Simple.Mountpoint + "/" + vmName
	diskFound := false
	for i, v := range vmConf.Disks {
		if v.DiskImage == "seed.iso" {
			return errors.New("CloudInit ISO has already been mounted")
		}

		if v.DiskImage == "seed-empty.iso" {
			diskFound = true
			vmConf.Disks[i].DiskImage = "seed.iso"
			vmConf.Disks[i].Comment = "CloudInit ISO file"

			err := HosterVmUtils.ConfigFileWriter(vmConf.VmConfig, vmFolder+"/"+HosterVmUtils.VM_CONFIG_NAME)
			if err != nil {
				return nil
			}

			if vmConf.Running {
				emojlog.PrintLogMessage("Please don't forget to reboot the VM to apply changes", emojlog.Debug)
			}
			emojlog.PrintLogMessage("CloudInit ISO has been mounted", emojlog.Changed)

			return nil
		}
	}

	if !diskFound {
		return errors.New("CloudInit ISO disk could not be found")
	}
	return nil
}

func unmountCiIso(vmName string) error {
	fileContents := []byte("placeholder file for an empty CI ISO file")

	vmConf, err := HosterVmUtils.InfoJsonApi(vmName)
	if err != nil {
		return err
	}

	diskFound := false
	vmFolder := vmConf.Simple.Mountpoint + "/" + HosterVmUtils.VM_CONFIG_NAME
	for i, v := range vmConf.Disks {
		if v.DiskImage == "seed-empty.iso" {
			return errors.New("CloudInit ISO has already been unmounted")
		}

		if v.DiskImage == "seed.iso" {
			diskFound = true
			vmConf.Disks[i].DiskImage = "seed-empty.iso"
			vmConf.Disks[i].Comment = "An empty CloudInit ISO file"

			err := HosterVmUtils.ConfigFileWriter(vmConf.VmConfig, vmFolder+"/"+HosterVmUtils.VM_CONFIG_NAME)
			if err != nil {
				return nil
			}

			if vmConf.Running {
				emojlog.PrintLogMessage("Please don't forget to reboot the VM to apply changes", emojlog.Debug)
			}
			emojlog.PrintLogMessage("CloudInit ISO has been unmounted", emojlog.Changed)

			return nil
		}
	}
	if !diskFound {
		return errors.New("CloudInit ISO disk could not be found")
	}

	err = os.WriteFile(vmFolder+"/placeholder", fileContents, 0640)
	if err != nil {
		return nil
	}

	_ = os.Remove(vmFolder + "/seed-empty.iso")
	out, err := exec.Command("genisoimage", "-output", vmFolder+"/seed-empty.iso", "-volid", "cidata", "-joliet", "-rock", vmFolder+"/placeholder").CombinedOutput()
	if err != nil {
		return errors.New("there was a problem generating an ISO: " + string(out) + "; " + err.Error())
	}
	_ = os.Remove(vmFolder + "/placeholder")

	return nil
}
