package HosterVmUtils

import (
	FileExists "HosterCore/internal/pkg/file_exists"
	"errors"
	"fmt"
	"os/exec"
	"strings"
)

func DiskExpandOffline(diskImage string, expansionSize int, vmName string) error {
	vm, err := InfoJsonApi(vmName)
	if err != nil {
		return err
	}

	if vm.Running {
		return errors.New("vm has to be offline, due to the data loss possibility of online expansion")
	}
	if vm.Backup {
		return errors.New("this is a backup VM")
	}
	if len(diskImage) < 1 {
		return errors.New("disk image name is empty")
	}
	if expansionSize < 1 {
		return errors.New("expansion size is less than 1G")
	}

	diskImageLocation := ""
	if strings.Contains(diskImage, "/") {
		diskImageLocation = diskImage
	} else {
		vmFolder := vm.Simple.Mountpoint + "/" + vm.Name
		diskImageLocation = vmFolder + "/" + diskImage
	}

	if !FileExists.CheckUsingOsStat(diskImageLocation) {
		return errors.New("disk image doesn't exist: " + diskImageLocation)
	}

	expansionSizeStr := "+" + fmt.Sprintf("%d", expansionSize) + "G"
	out, err := exec.Command("truncate", "-s", expansionSizeStr, diskImageLocation).CombinedOutput()
	if err != nil {
		return errors.New("can't expand the drive: " + strings.TrimSpace(string(out)) + "; " + strings.TrimSpace(err.Error()))
	}

	return nil
}
