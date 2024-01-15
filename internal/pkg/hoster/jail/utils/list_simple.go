package HosterJailUtils

import (
	HosterHost "HosterCore/internal/pkg/hoster/host"
	HosterZfs "HosterCore/internal/pkg/hoster/zfs"
	"fmt"
	"os"
)

type JailListSimple struct {
	JailName string
	HosterZfs.MountPoint
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

	return
}
