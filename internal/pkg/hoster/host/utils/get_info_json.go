// Copyright 2023 Hoster Authors. All rights reserved.
// Use of this source code is governed by an Apache License 2.0
// license that can be found in the LICENSE file.

package HosterHostUtils

import (
	FreeBSDOsInfo "HosterCore/internal/pkg/freebsd/info"
	FreeBSDPgrep "HosterCore/internal/pkg/freebsd/pgrep"
	FreeBSDsysctls "HosterCore/internal/pkg/freebsd/sysctls"
	HosterLocations "HosterCore/internal/pkg/hoster/locations"
	HosterVmUtils "HosterCore/internal/pkg/hoster/vm/utils"
	timeconversion "HosterCore/internal/pkg/time_conversion"
	zfsutils "HosterCore/internal/pkg/zfs_utils"
	"os/exec"
	"slices"
	"strings"
	"sync"
)

type HosterServices struct {
	DnsServerRunning    bool   `json:"dns_server_running"`
	SchedulerRunning    bool   `json:"scheduler_running"`
	RestApiRunning      bool   `json:"rest_api_running"`
	NodeExporterRunning bool   `json:"node_exporter_running"`
	HaWatchdogRunning   bool   `json:"ha_watchdog_running"`
	DnsServerPID        int    `json:"dns_server_pid"`
	SchedulerPID        int    `json:"scheduler_pid"`
	RestApiPID          int    `json:"rest_api_pid"`
	NodeExporterPID     int    `json:"node_exporter_pid"`
	HaWatchdogPID       int    `json:"ha_watchdog_pid"`
	DnsServerVersion    string `json:"dns_server_version"`
	SchedulerVersion    string `json:"scheduler_version"`
	RestApiVersion      string `json:"rest_api_version"`
	NodeExporterVersion string `json:"node_exporter_version"`
	HaWatchdogVersion   string `json:"ha_watchdog_version"`
	HosterVersion       string `json:"hoster_version"`
	VmSupervisorVersion string `json:"vm_supervisor_version"`
	MBufferVersion      string `json:"mbuffer_version"`
	SelfUpdateVersion   string `json:"self_update_version"`
}

type HostInfo struct {
	Services           HosterServices         `json:"services"`
	CpuInfo            FreeBSDOsInfo.CpuInfo  `json:"cpu_info"`
	RamInfo            FreeBSDOsInfo.RamInfo  `json:"ram_info"`
	SwapInfo           FreeBSDOsInfo.SwapInfo `json:"swap_info"`
	ArcInfo            FreeBSDOsInfo.ArcInfo  `json:"arc_info"`
	ZpoolList          []zfsutils.ZpoolInfo   `json:"zpool_list"`
	VCPU2PCURatio      float64                `json:"vcpu_2_pcpu_ratio"`
	AllVms             int                    `json:"all_vms"`
	LiveVms            int                    `json:"live_vms"`
	BackupVms          int                    `json:"backup_vms"`
	OfflineVms         int                    `json:"offline_vms"`
	OfflineVmsProd     int                    `json:"offline_vms_prod"`
	VCPU2PCU           string                 `json:"-"`
	Hostname           string                 `json:"hostname"`
	SystemUptime       string                 `json:"system_uptime"`
	SystemMajorVersion string                 `json:"system_major_version"`
	RunningKernel      string                 `json:"running_kernel"`
	LatestKernel       string                 `json:"latest_kernel"`
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
				if conf.Production {
					r.OfflineVmsProd += 1
				}
			} else {
				r.BackupVms += 1
			}
		}
		r.VCPU2PCU, r.VCPU2PCURatio = GetPc2VcRatioLazy(cpusUsed)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		bootTime, err := FreeBSDsysctls.SysctlKernBoottime()
		if err != nil {
			r.SystemUptime = "0s"
		}
		r.SystemUptime = timeconversion.UnixTimeToUptime(bootTime.Sec)

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

		infoArc, err := FreeBSDOsInfo.GetArcInfo()
		if err == nil {
			r.ArcInfo = infoArc
		}

		infoZpoolList, err := zfsutils.GetZpoolList()
		if err == nil {
			r.ZpoolList = infoZpoolList
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()

		infoCpu, err := FreeBSDOsInfo.GetCpuInfo()
		if err == nil {
			r.CpuInfo = infoCpu
		}

		infoSwap, err := FreeBSDOsInfo.GetSwapInfo()
		if err == nil {
			r.SwapInfo = infoSwap
		}

		infoRam, err := FreeBSDOsInfo.GetRamInfo()
		if err == nil {
			r.RamInfo = infoRam
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()

		pid, err := FreeBSDPgrep.FindDnsServer()
		if err == nil {
			r.Services.DnsServerRunning = true
			r.Services.DnsServerPID = pid
		}

		pid, err = FreeBSDPgrep.FindNodeExporter()
		if err == nil {
			r.Services.NodeExporterPID = pid
			r.Services.NodeExporterRunning = true
		}

		pid, err = FreeBSDPgrep.FindScheduler()
		if err == nil {
			r.Services.SchedulerPID = pid
			r.Services.SchedulerRunning = true
		}

		pid, err = FreeBSDPgrep.FindWatchdog()
		if err == nil {
			r.Services.HaWatchdogPID = pid
			r.Services.HaWatchdogRunning = true
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()

		// Hoster
		binary, err := HosterLocations.LocateBinary("hoster")
		if err == nil {
			out, err := exec.Command(binary, "version").CombinedOutput()
			if err == nil {
				r.Services.HosterVersion = strings.TrimSpace(string(out))
			}
		}

		// VM Supervisor
		binary, err = HosterLocations.LocateBinary("vm_supervisor_service")
		if err == nil {
			out, err := exec.Command(binary, "version").CombinedOutput()
			if err == nil {
				r.Services.VmSupervisorVersion = strings.TrimSpace(string(out))
			}
		}

		// DNS Server
		binary, err = HosterLocations.LocateBinary("dns_server")
		if err == nil {
			out, err := exec.Command(binary, "version").CombinedOutput()
			if err == nil {
				r.Services.DnsServerVersion = strings.TrimSpace(string(out))
			}
		}

		// HA Watchdog
		binary, err = HosterLocations.LocateBinary("ha_watchdog")
		if err == nil {
			out, err := exec.Command(binary, "version").CombinedOutput()
			if err == nil {
				r.Services.HaWatchdogVersion = strings.TrimSpace(string(out))
			}
		}

		// Scheduler
		binary, err = HosterLocations.LocateBinary("scheduler")
		if err == nil {
			out, err := exec.Command(binary, "version").CombinedOutput()
			if err == nil {
				r.Services.SchedulerVersion = strings.TrimSpace(string(out))
			}
		}

		// Self Update
		binary, err = HosterLocations.LocateBinary("self_update")
		if err != nil {
			out, err := exec.Command(binary, "version").CombinedOutput()
			if err != nil {
				r.Services.SelfUpdateVersion = strings.TrimSpace(string(out))
			}
		}

		// MBuffer
		binary, err = HosterLocations.LocateBinary("mbuffer")
		if err == nil {
			out, err := exec.Command(binary, "version").CombinedOutput()
			if err == nil {
				r.Services.MBufferVersion = strings.TrimSpace(string(out))
			}
		}

		// Node Exporter
		binary, err = HosterLocations.LocateBinary("node_exporter_custom")
		if err == nil {
			out, err := exec.Command(binary, "version").CombinedOutput()
			if err == nil {
				r.Services.NodeExporterVersion = strings.TrimSpace(string(out))
			}
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()

		out, err := exec.Command("uname", "-r").CombinedOutput()
		if err == nil {
			r.RunningKernel = strings.TrimSpace(string(out))
		}

		out, err = exec.Command("freebsd-version", "-k").CombinedOutput()
		if err == nil {
			r.LatestKernel = strings.TrimSpace(string(out))
		}
	}()

	wg.Wait()
	return
}
