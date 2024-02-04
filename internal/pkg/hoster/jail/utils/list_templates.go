package HosterJailUtils

import (
	HosterHost "HosterCore/internal/pkg/hoster/host"
	HosterZfs "HosterCore/internal/pkg/hoster/zfs"
	"os"
	"regexp"
)

func ListTemplates() (r []string, e error) {
	reMatchTemplate := regexp.MustCompile(`/jail-template-`)

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
			if file.IsDir() && reMatchTemplate.MatchString(file.Name()) {
				r = append(r, v.DsName+"/"+file.Name())
			}
		}
	}

	return
}
