// Copyright 2024 Hoster Authors. All rights reserved.
// Use of this source code is governed by an Apache License 2.0
// license that can be found in the LICENSE file.

package HosterVmUtils

import (
	HosterHost "HosterCore/internal/pkg/hoster/host"
	HosterZfs "HosterCore/internal/pkg/hoster/zfs"
	"fmt"
	"os"
)

type VmListSimple struct {
	VmName string
	HosterZfs.MountPoint
}

// Scans all Hoster-related ZFS datasets in search for the VM config files.
//
// Returns a list of VMs found + their basic ZFS dataset parameters (check the struct for the list of such parameters).
func ListAllSimple() (r []VmListSimple, e error) {
	// Get the host config
	hostConfig, err := HosterHost.GetHostConfig()
	if err != nil {
		e = err
		return
	}
	// Get the list of all existing ZFS datasets+mountpoints
	mountPoints, err := HosterZfs.ListMountPoints()
	if err != nil {
		e = err
		return
	}
	// Check if the dataset from host config is present, and append it's ds+mountpoint to the slice
	var mpsToScan []HosterZfs.MountPoint
	for _, v := range hostConfig.ActiveZfsDatasets {
		for _, vv := range mountPoints {
			if v == vv.DsName {
				mpsToScan = append(mpsToScan, vv)
			}
		}
	}
	// Compile the list of VMs present on this system
	for _, v := range mpsToScan {
		// Get the list of all files and directories on a given ZFS mountpoint
		files, err := os.ReadDir(v.Mountpoint)
		if err != nil {
			continue // Skip the folder if we could not read it (whatever the reason is - we don't really care)
		}
		// Iterate over folders, and locate VM config files (if the config was found, add the VM to our final list)
		for _, file := range files {
			if file.IsDir() && VmConfigExists(fmt.Sprintf("%s/%s", v.Mountpoint, file.Name())) {
				vm := VmListSimple{}
				vm.MountPoint = v
				vm.VmName = file.Name()
				r = append(r, vm)
			}
		}
	}

	return
}
