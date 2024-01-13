package HosterJail

import (
	HosterHost "HosterCore/pkg/hoster/host"
	HosterZfs "HosterCore/pkg/hoster/zfs"
	"fmt"
	"os"
)

type JailListSimple struct {
	Name string
	HosterZfs.MountPoint
}

func ListAllSimple() (r []JailListSimple, e error) {
	hostConfig, err := HosterHost.GetHostConfig()
	if err != nil {
		e = err
		return
	}

	mountPoints, err := HosterZfs.ZfsListMountPoints()
	if err != nil {
		e = err
		return
	}

	var mpsToScan []HosterZfs.MountPoint
	for _, v := range hostConfig.ActiveZfsDatasets {
		for _, vv := range mountPoints {
			if v == vv.Name {
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
			if file.IsDir() && JailExists(fmt.Sprintf("%s/%s", v.Mountpoint, file.Name())) {
				r = append(r, JailListSimple{MountPoint: v, Name: file.Name()})
			}
		}
	}

	return
}
