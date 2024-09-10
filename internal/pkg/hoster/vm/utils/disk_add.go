package HosterVmUtils

import (
	FileExists "HosterCore/internal/pkg/file_exists"
	"fmt"
	"os/exec"
	"slices"
	"strings"
)

func AddNewVmDisk(vmName string, input VmDisk) error {
	validDrivers := []string{"ahci-hd", "virtio-blk", "nvme"}
	if !slices.Contains(validDrivers, input.DiskType) {
		return fmt.Errorf("invalid disk type")
	}

	validLocations := []string{"external", "internal"}
	if !slices.Contains(validLocations, input.DiskLocation) {
		return fmt.Errorf("invalid disk location")
	}

	if len(input.DiskImage) < 1 && input.DiskLocation == "external" {
		return fmt.Errorf("disk file path cannot be empty")
	}

	if !strings.HasSuffix(strings.ToLower(input.DiskImage), ".img") {
		if input.DiskLocation == "external" {
			return fmt.Errorf("disk image file must have an .img extension")
		}
	}

	if len(input.Comment) < 1 {
		input.Comment = "Additional data disk" // default comment
	}
	if input.DiskInputSize < 1 {
		input.DiskInputSize = 10 // if the size is not set, default to 10GB
	}

	vmInfo, err := InfoJsonApi(vmName)
	if err != nil {
		return err
	}

	if input.DiskLocation == "internal" {
		// Example location this call should generate: /tank/vm-encrypted/test-vm-1/disk0.img
		input.DiskImage = vmInfo.Simple.Mountpoint + "/" + vmName + "/disk" + fmt.Sprintf("%d", len(vmInfo.VmConfig.Disks)) + ".img"
	}
	if FileExists.CheckUsingOsStat(input.DiskImage) {
		return fmt.Errorf("disk file already exists")
	}

	for _, v := range vmInfo.VmConfig.Disks {
		if v.DiskImage == input.DiskImage {
			return fmt.Errorf("disk image file is already mounted")
		}
	}

	out, err := exec.Command("truncate", "-s", fmt.Sprintf("%dG", input.DiskInputSize), input.DiskImage).CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s; %s", strings.TrimSpace(string(out)), strings.TrimSpace(err.Error()))
	}

	input.DiskInputSize = 0
	if input.DiskLocation == "internal" {
		split := strings.Split(input.DiskImage, "/")
		input.DiskImage = split[len(split)-1]
	}
	vmInfo.VmConfig.Disks = append(vmInfo.VmConfig.Disks, input)

	configLocation := vmInfo.Simple.Mountpoint + "/" + vmName + "/" + VM_CONFIG_NAME
	err = ConfigFileWriter(vmInfo.VmConfig, configLocation)
	if err != nil {
		return err
	}

	return nil
}
