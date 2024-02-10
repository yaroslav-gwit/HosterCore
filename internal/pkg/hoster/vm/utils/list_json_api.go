// Copyright 2023 Hoster Authors. All rights reserved.
// Use of this source code is governed by an Apache License 2.0
// license that can be found in the LICENSE file.

package HosterVmUtils

import (
	FreeBSDps "HosterCore/internal/pkg/freebsd/ps"
	FreeBSDsysctls "HosterCore/internal/pkg/freebsd/sysctls"
	timeconversion "HosterCore/internal/pkg/time_conversion"
	"fmt"
	"regexp"
	"slices"
)

type DiskInfoApi struct {
	TotalBytes uint64
	TotalHuman string
	UsedBytes  uint64
	UsedHuman  string
}

type VmApi struct {
	VmConfig
	Name        string        `json:"name"`
	Uptime      string        `json:"uptime"`
	UptimeUnix  int64         `json:"uptime_unix"`
	Running     bool          `json:"running"`
	Backup      bool          `json:"backup"`
	Production  bool          `json:"production"`
	Encrypted   bool          `json:"encrypted"`
	CurrentHost string        `json:"current_host"`
	DiskInfo    []DiskInfoApi `json:"disk_info"`
}

func ListJsonApi() (r []VmApi, e error) {
	ps, err := FreeBSDps.ProcessTimes()
	if err != nil {
		e = err
		return
	}

	vms, err := ListAllSimple()
	if err != nil {
		e = err
		return
	}
	liveVms, err := GetRunningVms()
	if err != nil {
		e = err
		return
	}
	hostname, _ := FreeBSDsysctls.SysctlKernHostname()

	for _, v := range vms {
		temp := VmApi{}
		conf, err := GetVmConfig(v.Mountpoint + "/" + v.VmName)
		if err != nil {
			fmt.Println(err)
			continue
		}

		temp.Name = v.VmName
		temp.VmConfig = conf
		temp.Encrypted = v.Encrypted

		if slices.Contains(liveVms, v.VmName) {
			temp.Running = true
			reMatchVmProcess := regexp.MustCompile(`bhyve:\s+` + v.VmName + `($|\s+)`)
			for _, vv := range ps {
				if reMatchVmProcess.MatchString(vv.Command) {
					temp.Uptime = timeconversion.ProcessUptimeToHuman(vv.StartTime)
					temp.UptimeUnix = vv.StartTime
					break
				}
			}
		} else {
			temp.Uptime = "0s"
		}

		temp.CurrentHost = hostname
		if hostname != conf.ParentHost {
			temp.Backup = true
		}

		if conf.LiveStatus == "prod" || conf.LiveStatus == "production" {
			temp.Production = true
		}

		diskInfo, err := DiskInfo(v.Mountpoint + "/" + v.VmName + "/disk0.img")
		if err != nil {
			continue
		}
		temp.DiskInfo = append(temp.DiskInfo, diskInfo)
	}

	return
}
