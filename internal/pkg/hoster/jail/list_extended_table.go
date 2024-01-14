package HosterJail

import (
	"HosterCore/internal/pkg/byteconversion"
	FreeBSDsysctls "HosterCore/internal/pkg/freebsd/sysctls"
	HosterZfs "HosterCore/internal/pkg/hoster/zfs"
	"fmt"
)

type JailListExtendedTable struct {
	Name             string
	Status           string
	CPULimit         string
	RAMLimit         string
	MainIpAddress    string
	Release          string
	Uptime           string
	StorageUsed      string
	StorageAvailable string
	Description      string
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
		jailDsFolder := v.MountPoint.DsName + "/" + v.JailName

		jailConfig, err := GetJailConfig(jailDsFolder)
		if err != nil {
			fmt.Println(err)
			continue
		}

		jailStruct.Name = v.JailName

		if jailConfig.Parent == hostname {
			for _, vv := range onlineJails {
				if v.JailName == vv.Name {
					if vv.Running {
						jailStruct.Status += JAIL_EMOJI_ONLINE
					} else {
						jailStruct.Status += JAIL_EMOJI_OFFLINE
					}
				}
			}
		} else {
			jailStruct.Status += JAIL_EMOJI_BACKUP
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
			fmt.Println(err)
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
