// Copyright 2023 Hoster Authors. All rights reserved.
// Use of this source code is governed by an Apache License 2.0
// license that can be found in the LICENSE file.

package HosterHostUtils

import (
	FreeBSDOsInfo "HosterCore/internal/pkg/freebsd/info"
	FreeBSDsysctls "HosterCore/internal/pkg/freebsd/sysctls"
	HosterVmUtils "HosterCore/internal/pkg/hoster/vm/utils"
	timeconversion "HosterCore/internal/pkg/time_conversion"
	zfsutils "HosterCore/internal/pkg/zfs_utils"
	"slices"
	"sync"
)

type HostInfo struct {
	AllVms             int                    `json:"all_vms"`
	LiveVms            int                    `json:"live_vms"`
	BackupVms          int                    `json:"backup_vms"`
	OfflineVms         int                    `json:"offline_vms"`
	OfflineVmsProd     int                    `json:"offline_vms_prod"`
	VCPU2PCURatio      float64                `json:"vcpu_2_pcpu_ratio"`
	Hostname           string                 `json:"hostname"`
	SystemUptime       string                 `json:"system_uptime"`
	SystemMajorVersion string                 `json:"system_major_version"`
	CpuInfo            FreeBSDOsInfo.CpuInfo  `json:"cpu_info"`
	RamInfo            FreeBSDOsInfo.RamInfo  `json:"ram_info"`
	SwapInfo           FreeBSDOsInfo.SwapInfo `json:"swap_info"`
	ArcInfo            FreeBSDOsInfo.ArcInfo  `json:"arc_info"`
	ZpoolList          []zfsutils.ZpoolInfo   `json:"zpool_list"`
}

func GetHostInfo() (r HostInfo, e error) {
	var wg = &sync.WaitGroup{}
	r.Hostname, _ = FreeBSDsysctls.SysctlKernHostname()

	wg.Add(1)
	go func() {
		defer wg.Done()
		cpusUsed := 0
		vms, err := HosterVmUtils.ListAllSimple()
		if err != nil {
			return
		}
		list, err := HosterVmUtils.GetRunningVms()
		if err != nil {
			return
		}

		for _, v := range vms {
			conf, err := HosterVmUtils.GetVmConfig(v.Mountpoint + "/" + v.VmName)
			if err != nil {
				continue
			}

			r.AllVms += 1
			if slices.Contains(list, v.VmName) {
				r.LiveVms += 1
				cpusUsed = cpusUsed + (r.CpuInfo.Sockets * r.CpuInfo.Cores * r.CpuInfo.Threads)
			} else if conf.ParentHost == r.Hostname {
				r.OfflineVms += 1
				if conf.LiveStatus == "prod" || conf.LiveStatus == "production" {
					r.OfflineVmsProd += 1
				}
			} else {
				r.BackupVms += 1
			}
		}
		_, r.VCPU2PCURatio = GetPc2VcRatioLazy(cpusUsed)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		bootTime, err := FreeBSDsysctls.SysctlKernBoottime()
		if err != nil {
			r.SystemUptime = "0s"
		}
		r.SystemUptime = timeconversion.KernBootToUptime(bootTime.USec)

		ver, err := FreeBSDOsInfo.GetMajorReleaseVersion()
		if err != nil {
			ver = "NULL"
		} else {
			r.SystemMajorVersion = ver
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		info, err := FreeBSDOsInfo.GetRamInfo()
		if err != nil {
			return
		}
		r.RamInfo = info
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		info, err := FreeBSDOsInfo.GetArcInfo()
		if err != nil {
			return
		}
		r.ArcInfo = info
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		info, err := zfsutils.GetZpoolList()
		if err != nil {
			return
		}
		r.ZpoolList = info
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		info, err := FreeBSDOsInfo.GetCpuInfo()
		if err != nil {
			return
		}
		r.CpuInfo = info
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		info, err := FreeBSDOsInfo.GetSwapInfo()
		if err != nil {
			return
		}
		r.SwapInfo = info
	}()

	wg.Wait()
	return
}
