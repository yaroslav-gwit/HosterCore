package HosterJailUtils

import (
	"HosterCore/internal/pkg/byteconversion"
	FreeBSDsysctls "HosterCore/internal/pkg/freebsd/sysctls"
	HosterZfs "HosterCore/internal/pkg/hoster/zfs"
	"fmt"
)

type JailApi struct {
	JailConfig
	Simple         JailListSimple `json:"-"`
	Name           string         `json:"name"`
	Uptime         string         `json:"uptime"`
	Running        bool           `json:"running"`
	Encrypted      bool           `json:"encrypted"`
	Backup         bool           `json:"backup"`
	Release        string         `json:"release"`
	CurrentHost    string         `json:"current_host"`
	SpaceUsedHuman string         `json:"space_used_h"`
	SpaceUsedBytes uint64         `json:"space_used_b"`
	SpaceFreeHuman string         `json:"space_free_h"`
	SpaceFreeBytes uint64         `json:"space_free_b"`
}

func InfoJsonApi(jailName string) (r JailApi, e error) {
	hostname, _ := FreeBSDsysctls.SysctlKernHostname()
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

	for _, v := range jails {
		if v.JailName != jailName {
			continue
		}

		jailDsFolder := v.MountPoint.Mountpoint + "/" + v.JailName
		jailConfig, err := GetJailConfig(jailDsFolder)
		if err != nil {
			continue
		}

		r.Name = v.JailName
		r.JailConfig = jailConfig
		r.CurrentHost = hostname
		if jailConfig.Parent == hostname {
			for _, vv := range onlineJails {
				if v.JailName == vv.Name {
					r.Running = true
				}
			}
		} else {
			r.Backup = true
		}

		if v.MountPoint.Encrypted {
			r.Encrypted = true
		}

		release, err := ReleaseVersion(jailDsFolder)
		if err != nil {
			continue
		}
		r.Release = release
		r.Uptime = GetUptimeHuman(v.JailName)

		for _, vv := range zfsSpace {
			if v.MountPoint.DsName+"/"+v.JailName == vv.Name {
				r.SpaceUsedHuman = byteconversion.BytesToHuman(vv.Used)
				r.SpaceFreeHuman = byteconversion.BytesToHuman(vv.Available)
				r.SpaceUsedBytes = vv.Used
				r.SpaceFreeBytes = vv.Available
			}
		}

		return
	}

	e = fmt.Errorf("jail was not found")
	return
}
