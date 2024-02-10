// Copyright 2023 Hoster Authors. All rights reserved.
// Use of this source code is governed by an Apache License 2.0
// license that can be found in the LICENSE file.

package HosterJailUtils

import (
	"HosterCore/internal/pkg/byteconversion"
	FreeBSDsysctls "HosterCore/internal/pkg/freebsd/sysctls"
	HosterZfs "HosterCore/internal/pkg/hoster/zfs"
	"fmt"
)

type JailListExtendedTable struct {
	Name             string `json:"jail_name"` // Jail Name
	Running          bool   `json:"running"`   // is Jail Online/Running
	Status           string `json:"status"`    // status inside of the CLI table, e.g. 🟢🔒🔁
	CPULimit         string `json:"cpu_limit"` // CPU limit, e.g. 50%
	RAMLimit         string `json:"ram_limit"` // RAM limit, e.g. 10G
	MainIpAddress    string `json:"main_ip_address"`
	Release          string `json:"release"`
	Uptime           string `json:"uptime"` // Human readable uptime format, e.g. 20d 10h 19m 1s
	StorageUsed      string `json:"storage_used"`
	StorageAvailable string `json:"storage_available"`
	Description      string `json:"description"`
}

const JAIL_EMOJI_ONLINE = "🟢"
const JAIL_EMOJI_OFFLINE = "🔴"
const JAIL_EMOJI_BACKUP = "💾"
const JAIL_EMOJI_ENCRYPTED = "🔒"
const JAIL_EMOJI_PRODUCTION = "🔁"

func ListAllExtendedTable() (r []JailListExtendedTable, e error) {
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
		jailStruct := JailListExtendedTable{}
		jailDsFolder := v.MountPoint.Mountpoint + "/" + v.JailName

		jailConfig, err := GetJailConfig(jailDsFolder)
		if err != nil {
			// fmt.Println(err)
			continue
		}

		jailStruct.Name = v.JailName

		if jailConfig.Parent == hostname {
			for _, vv := range onlineJails {
				if v.JailName == vv.Name {
					jailStruct.Running = true
				}
			}
		} else {
			jailStruct.Status += JAIL_EMOJI_BACKUP
		}

		if jailStruct.Running {
			jailStruct.Status += JAIL_EMOJI_ONLINE
		} else {
			jailStruct.Status += JAIL_EMOJI_OFFLINE
		}

		if v.MountPoint.Encrypted {
			jailStruct.Status += JAIL_EMOJI_ENCRYPTED
		}

		if jailConfig.Production {
			jailStruct.Status += JAIL_EMOJI_PRODUCTION
		}

		jailStruct.CPULimit = fmt.Sprintf("%d%%", jailConfig.CPULimitPercent)
		jailStruct.RAMLimit = jailConfig.RAMLimit
		jailStruct.MainIpAddress = jailConfig.IPAddress

		release, err := ReleaseVersion(jailDsFolder)
		if err != nil {
			// fmt.Println(err)
			continue
		}
		jailStruct.Release = release
		jailStruct.Uptime = GetUptimeHuman(v.JailName)
		jailStruct.Description = jailConfig.Description

		for _, vv := range zfsSpace {
			if v.MountPoint.DsName+"/"+v.JailName == vv.Name {
				jailStruct.StorageUsed = byteconversion.BytesToHuman(vv.Used)
				jailStruct.StorageAvailable = byteconversion.BytesToHuman(vv.Available)
			}
		}

		r = append(r, jailStruct)
	}

	return
}
