// Copyright 2023 Hoster Authors. All rights reserved.
// Use of this source code is governed by an Apache License 2.0
// license that can be found in the LICENSE file.

package HosterJailUtils

import (
	HosterHost "HosterCore/internal/pkg/hoster/host"
	HosterZfs "HosterCore/internal/pkg/hoster/zfs"
	"fmt"
	"os"
	"sort"
	"strings"
)

type JailListSimple struct {
	JailName             string `json:"jail_name"`
	HosterZfs.MountPoint `json:"mount_point"`
}

// Scans all Hoster-related ZFS datasets in search for Jail config files.
//
// Returns a list of Jails found + their basic ZFS dataset parameters (check the struct for the list of such parameters).
func ListAllSimple() (r []JailListSimple, e error) {
	hostConfig, err := HosterHost.GetHostConfig()
	if err != nil {
		e = err
		return
	}

	mountPoints, err := HosterZfs.ListMountPoints()
	if err != nil {
		e = err
		return
	}

	var mpsToScan []HosterZfs.MountPoint
	for _, v := range hostConfig.ActiveZfsDatasets {
		for _, vv := range mountPoints {
			if v == vv.DsName {
				mpsToScan = append(mpsToScan, vv)
			}
		}
	}

	for _, v := range mpsToScan {
		files, err := os.ReadDir(v.Mountpoint)
		if err != nil {
			continue
		}

		for _, file := range files {
			if file.IsDir() && JailConfigExists(fmt.Sprintf("%s/%s", v.Mountpoint, file.Name())) {
				jailSimple := JailListSimple{}
				jailSimple.MountPoint = v
				jailSimple.JailName = file.Name()
				r = append(r, jailSimple)
			}
		}
	}

	sort.SliceStable(r, func(i, j int) bool {
		return strings.ToLower(r[i].JailName) < strings.ToLower(r[j].JailName)
	})

	return
}
