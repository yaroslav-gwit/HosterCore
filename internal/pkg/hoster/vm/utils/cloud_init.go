package HosterVmUtils

import (
	"errors"
	"os"
	"os/exec"
)

func MountCiIso(vmName string) error {
	vmConf, err := InfoJsonApi(vmName)
	if err != nil {
		return err
	}

	diskFound := false
	vmFolder := vmConf.Simple.Mountpoint + "/" + vmName

	for i, v := range vmConf.Disks {
		if v.DiskImage == "seed.iso" {
			return errors.New("CloudInit ISO has already been mounted")
		}

		if v.DiskImage == "seed-empty.iso" {
			diskFound = true
			vmConf.Disks[i].DiskImage = "seed.iso"
			vmConf.Disks[i].Comment = "CloudInit ISO file"

			err := ConfigFileWriter(vmConf.VmConfig, vmFolder+"/"+VM_CONFIG_NAME)
			if err != nil {
				return nil
			}

			// if vmConf.Running {
			// 	emojlog.PrintLogMessage("Please don't forget to reboot the VM to apply changes", emojlog.Debug)
			// }
			// emojlog.PrintLogMessage("CloudInit ISO has been mounted", emojlog.Changed)

			return nil
		}
	}

	if !diskFound {
		return errors.New("CloudInit ISO disk could not be found")
	}

	return nil
}

func UnmountCiIso(vmName string) error {
	fileContents := []byte("placeholder file for an empty CI ISO file")

	vmConf, err := InfoJsonApi(vmName)
	if err != nil {
		return err
	}

	diskFound := false
	vmFolder := vmConf.Simple.Mountpoint + "/" + VM_CONFIG_NAME
	for i, v := range vmConf.Disks {
		if v.DiskImage == "seed-empty.iso" {
			return errors.New("CloudInit ISO has already been unmounted")
		}

		if v.DiskImage == "seed.iso" {
			diskFound = true
			vmConf.Disks[i].DiskImage = "seed-empty.iso"
			vmConf.Disks[i].Comment = "An empty CloudInit ISO file"

			err := ConfigFileWriter(vmConf.VmConfig, vmFolder+"/"+VM_CONFIG_NAME)
			if err != nil {
				return nil
			}

			// if vmConf.Running {
			// 	emojlog.PrintLogMessage("Please don't forget to reboot the VM to apply changes", emojlog.Debug)
			// }
			// emojlog.PrintLogMessage("CloudInit ISO has been unmounted", emojlog.Changed)
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
