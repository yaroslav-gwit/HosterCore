package HosterVmUtils

import (
	FileExists "HosterCore/internal/pkg/file_exists"
	"fmt"
	"strings"
)

// This function mounts an ISO file inside a VM. This ISO is used for the OS installation purposes most of the time.
//
// The function returns an error if something goes wrong,
// otherwise it returns nil and writes the new config file with the absolute ISO file path in it.
func MountInstallationIso(vmName string, isoPath string, isoComment string) error {
	if len(isoComment) < 1 {
		isoComment = "Installation ISO file"
	}

	if len(isoPath) < 1 {
		return fmt.Errorf("ISO file path is empty")
	}

	if !strings.HasSuffix(strings.ToLower(isoPath), ".iso") {
		return fmt.Errorf("ISO file must have an .iso extension")
	}

	if !FileExists.CheckUsingOsStat(isoPath) {
		return fmt.Errorf("ISO file does not exist")
	}

	vmInfo, err := InfoJsonApi(vmName)
	if err != nil {
		return err
	}

	disk := VmDisk{}
	disk.Comment = isoComment
	disk.DiskImage = isoPath
	disk.DiskLocation = "external"
	disk.DiskType = "ahci-cd"

	for _, v := range vmInfo.VmConfig.Disks {
		if v.DiskImage == isoPath {
			return fmt.Errorf("ISO file is already mounted")
		}
	}
	vmInfo.VmConfig.Disks = append(vmInfo.VmConfig.Disks, disk)

	configLocation := vmInfo.Simple.Mountpoint + "/" + vmName + "/" + VM_CONFIG_NAME
	err = ConfigFileWriter(vmInfo.VmConfig, configLocation)
	if err != nil {
		return err
	}

	return nil
}

func UnmountInstallationIso(vmName string, isoPath string) error {
	if len(isoPath) < 1 {
		return fmt.Errorf("ISO file path is empty")
	}

	if !strings.HasSuffix(strings.ToLower(isoPath), ".iso") {
		return fmt.Errorf("ISO file must have an .iso extension")
	}

	vmInfo, err := InfoJsonApi(vmName)
	if err != nil {
		return err
	}

	disks := []VmDisk{}
	for _, v := range vmInfo.VmConfig.Disks {
		if v.DiskImage == isoPath {
			continue
		} else {
			disks = append(disks, v)
		}
	}

	if len(disks) == len(vmInfo.VmConfig.Disks) {
		return fmt.Errorf("ISO file is not mounted")
	}

	vmInfo.VmConfig.Disks = disks
	configLocation := vmInfo.Simple.Mountpoint + "/" + vmName + "/" + VM_CONFIG_NAME
	err = ConfigFileWriter(vmInfo.VmConfig, configLocation)
	if err != nil {
		return err
	}

	return nil
}
