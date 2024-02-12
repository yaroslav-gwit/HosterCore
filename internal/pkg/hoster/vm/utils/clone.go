// Copyright 2023 Hoster Authors. All rights reserved.
// Use of this source code is governed by an Apache License 2.0
// license that can be found in the LICENSE file.

package HosterVmUtils

import (
	zfsutils "HosterCore/internal/pkg/zfs_utils"
	"fmt"
	"os/exec"
	"strings"
)

func Clone(vmName string, newVmName string, snapshotName string) error {
	vms, err := ListAllSimple()
	if err != nil {
		return err
	}

	vmFound := false
	vmInfo := VmListSimple{}
	for _, v := range vms {
		if v.VmName == vmName {
			vmInfo = v
			vmFound = true
		}
	}
	if !vmFound {
		return fmt.Errorf("vm doesn't exist")
	}

	snaps, err := zfsutils.SnapshotListAll()
	if err != nil {
		return err
	}

	snapFound := false
	if len(snapshotName) < 1 {
		snapshotName = snaps[0].Name
		snapFound = true
	} else {
		for _, v := range snaps {
			if v.Name == snapshotName {
				snapFound = true
			}
		}
	}
	if !snapFound {
		return fmt.Errorf("snapshot doesn't exist")
	}

	out, err := exec.Command("zfs", "clone", snapshotName, vmInfo.DsName+"/"+newVmName).CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s; %s", strings.TrimSpace(string(out)), err.Error())
	}

	return nil
}
