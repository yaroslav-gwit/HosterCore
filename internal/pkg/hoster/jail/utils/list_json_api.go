// Copyright 2023 Hoster Authors. All rights reserved.
// Use of this source code is governed by an Apache License 2.0
// license that can be found in the LICENSE file.

package HosterJailUtils

import (
	"HosterCore/internal/pkg/byteconversion"
	FreeBSDsysctls "HosterCore/internal/pkg/freebsd/sysctls"
	HosterZfs "HosterCore/internal/pkg/hoster/zfs"
)

func ListJsonApi() (r []JailApi, e error) {
	jails, err := ListAllSimple()
	if err != nil {
		e = err
		return
	}

	onlineJails, err := GetRunningJails()
	if err != nil {
		e = err
		return
	}

	zfsSpace, err := HosterZfs.ListUsedAndAvailableSpace()
	if err != nil {
		e = err
		return
	}

	hostname, _ := FreeBSDsysctls.SysctlKernHostname()

	for _, v := range jails {
		jailStruct := JailApi{}
		jailDsFolder := v.MountPoint.Mountpoint + "/" + v.JailName

		jailConfig, err := GetJailConfig(jailDsFolder)
		if err != nil {
			continue
		}

		jailStruct.Name = v.JailName
		jailStruct.JailConfig = jailConfig
		jailStruct.CurrentHost = hostname
		jailStruct.Simple = v

		if jailConfig.Parent == hostname {
			for _, vv := range onlineJails {
				if v.JailName == vv.Name {
					jailStruct.Running = true
				}
			}
		} else {
			jailStruct.Backup = true
		}

		if v.MountPoint.Encrypted {
			jailStruct.Encrypted = true
		}

		release, err := ReleaseVersion(jailDsFolder)
		if err != nil {
			continue
		}
		jailStruct.Release = release
		jailStruct.Uptime = GetUptimeHuman(v.JailName)

		for _, vv := range zfsSpace {
			if v.MountPoint.DsName+"/"+v.JailName == vv.Name {
				jailStruct.SpaceUsedHuman = byteconversion.BytesToHuman(vv.Used)
				jailStruct.SpaceFreeHuman = byteconversion.BytesToHuman(vv.Available)
				jailStruct.SpaceUsedBytes = vv.Used
				jailStruct.SpaceFreeBytes = vv.Available
			}
		}
		r = append(r, jailStruct)
	}

	return
}
