// Copyright 2023 Hoster Authors. All rights reserved.
// Use of this source code is governed by an Apache License 2.0
// license that can be found in the LICENSE file.

package HosterJail

import (
	HosterJailUtils "HosterCore/internal/pkg/hoster/jail/utils"
	HosterVmUtils "HosterCore/internal/pkg/hoster/vm/utils"
	zfsutils "HosterCore/internal/pkg/zfs_utils"
	"fmt"
	"os/exec"
	"strings"
)

func Clone(jailName string, newJailName string, snapshotName string) error {
	// If the logger was already set, ignore this
	if !log.ConfigSet {
		log.SetFileLocation(HosterVmUtils.VM_AUDIT_LOG_LOCATION)
	}

	jails, err := HosterJailUtils.ListAllSimple()
	if err != nil {
		return err
	}

	jailFound := false
	jailInfo := HosterJailUtils.JailListSimple{}
	for _, v := range jails {
		if v.JailName == jailName {
			jailInfo = v
			jailFound = true
		}
	}
	if !jailFound {
		return fmt.Errorf("jail doesn't exist")
	}

	snaps, err := zfsutils.SnapshotListAll()
	if err != nil {
		return err
	}

	snapFound := false
	if len(snapshotName) < 1 {
		for _, v := range snaps {
			if jailInfo.DsName+"/"+jailName == v.Dataset {
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

	out, err := exec.Command("zfs", "clone", snapshotName, jailInfo.DsName+"/"+newJailName).CombinedOutput()
	if err != nil {
		errValue := fmt.Sprintf("%s; %s", strings.TrimSpace(string(out)), err.Error())
		log.Error("jail could not be cloned: " + jailName + "; error: " + errValue)
		return fmt.Errorf(errValue)
	}

	_, err = HosterJailUtils.WriteCache()
	if err != nil {
		return err
	}

	log.Warn("jail has been cloned: " + jailName + "; cloned jail name: " + newJailName)
	return nil
}
