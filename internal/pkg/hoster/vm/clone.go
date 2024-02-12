// Copyright 2023 Hoster Authors. All rights reserved.
// Use of this source code is governed by an Apache License 2.0
// license that can be found in the LICENSE file.

package HosterVm

import (
	HosterVmUtils "HosterCore/internal/pkg/hoster/vm/utils"
	zfsutils "HosterCore/internal/pkg/zfs_utils"
	"fmt"
	"os/exec"
	"strings"
)

func Clone(vmName string, newVmName string, snapshotName string) error {
	// If the logger was already set, ignore this
	if !log.ConfigSet {
		log.SetFileLocation(HosterVmUtils.VM_AUDIT_LOG_LOCATION)
	}

	vms, err := HosterVmUtils.ListAllSimple()
	if err != nil {
		return err
	}

	vmFound := false
	vmInfo := HosterVmUtils.VmListSimple{}
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
		for _, v := range snaps {
			if vmInfo.DsName+"/"+vmName == v.Dataset {
				snapFound = true
				snapshotName = v.Name
			}
		}
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
		errValue := fmt.Sprintf("%s; %s", strings.TrimSpace(string(out)), err.Error())
		log.Error("vm could not be cloned: " + vmName + "; error: " + errValue)
		return fmt.Errorf(errValue)
	}

	log.Warn("vm has been cloned: " + vmName + "; cloned vm name: " + newVmName)
	return nil
}
