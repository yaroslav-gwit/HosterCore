// Copyright 2023 Hoster Authors. All rights reserved.
// Use of this source code is governed by an Apache License 2.0
// license that can be found in the LICENSE file.

package HosterJailUtils

import (
	"HosterCore/internal/pkg/byteconversion"
	FreeBSDsysctls "HosterCore/internal/pkg/freebsd/sysctls"
	HosterZfs "HosterCore/internal/pkg/hoster/zfs"
)

type JailApi struct {
	JailConfig
	Name           string `json:"name"`
	Uptime         string `json:"uptime"`
	Running        bool   `json:"running"`
	Encrypted      bool   `json:"encrypted"`
	Backup         bool   `json:"backup"`
	Release        string `json:"release"`
	CurrentHost    string `json:"current_host"`
	SpaceUsedHuman string `json:"space_used_h"`
	SpaceUsedBytes uint64 `json:"space_used_b"`
	SpaceFreeHuman string `json:"space_free_h"`
	SpaceFreeBytes uint64 `json:"space_free_b"`
}

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
		jailStruct.CurrentHost = hostname
		jailStruct.Parent = jailConfig.Parent
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

		jailStruct.CPULimitPercent = jailConfig.CPULimitPercent
		jailStruct.RAMLimit = jailConfig.RAMLimit
		jailStruct.IPAddress = jailConfig.IPAddress

		release, err := ReleaseVersion(jailDsFolder)
		if err != nil {
			continue
		}
		jailStruct.Release = release
		jailStruct.Uptime = GetUptimeHuman(v.JailName)
		jailStruct.Description = jailConfig.Description

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
