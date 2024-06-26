// Copyright 2023 Hoster Authors. All rights reserved.
// Use of this source code is governed by an Apache License 2.0
// license that can be found in the LICENSE file.

package rctl

import (
	FreeBSDOsInfo "HosterCore/internal/pkg/freebsd/info"
	"fmt"
	"math"
	"os/exec"
	"strconv"
	"strings"
)

type RctMetrics struct {
	CpuTime      uint64 `json:"cpu_time"`
	DataSize     uint64 `json:"data_size"`
	StackSize    uint64 `json:"stack_size"`
	CoreDumpSize uint64 `json:"code_dump_size"`
	MemoryUse    uint64 `json:"memory_use"`
	MemoryLocked uint64 `json:"memory_locked"`
	MaxProc      int    `json:"max_proc"`
	OpenFiles    int    `json:"open_files"`
	VMemoryUse   uint64 `json:"vmemory_use"`
	NThr         int    `json:"nthr"`
	Nsemop       int    `json:"nsemop"`
	WallClock    int64  `json:"wall_clock"`
	PCpu         int    `json:"p_cpu"`
	ReadBps      uint64 `json:"read_bps"`
	WriteBps     uint64 `json:"write_bps"`
	ReadIoPs     uint64 `json:"read_iops"`
	WriteIoPs    uint64 `json:"write_iops"`
}

func MetricsProcess(pid int) (r RctMetrics, e error) {
	out, err := exec.Command("rctl", "-u", fmt.Sprintf("process:%d", pid)).CombinedOutput()
	if err != nil {
		e = fmt.Errorf("%s; %s", strings.TrimSpace(string(out)), err.Error())
		return
	}

	r, e = parseMetrics(string(out))
	return
}

func MetricsJail(jailName string) (r RctMetrics, e error) {
	out, err := exec.Command("rctl", "-u", fmt.Sprintf("jail:%s", jailName)).CombinedOutput()
	if err != nil {
		e = fmt.Errorf("%s; %s", strings.TrimSpace(string(out)), err.Error())
		return
	}

	r, e = parseMetrics(string(out))
	return
}

func parseMetrics(metrics string) (r RctMetrics, e error) {
	var err error

	for _, v := range strings.Split(metrics, "\n") {
		err = nil
		v = strings.TrimSpace(v)

		if strings.HasPrefix(v, CPU_TIME) {
			v = strings.TrimPrefix(v, CPU_TIME)
			r.CpuTime, err = strconv.ParseUint(v, 10, 64)
			if err != nil {
				e = err
				return
			}
			continue
		}

		if strings.HasPrefix(v, DATA_SIZE) {
			v = strings.TrimPrefix(v, DATA_SIZE)
			r.DataSize, err = strconv.ParseUint(v, 10, 64)
			if err != nil {
				e = err
				return
			}
			continue
		}

		if strings.HasPrefix(v, STACK_SIZE) {
			v = strings.TrimPrefix(v, STACK_SIZE)
			r.StackSize, err = strconv.ParseUint(v, 10, 64)
			if err != nil {
				e = err
				return
			}
			continue
		}

		if strings.HasPrefix(v, CORE_DUMP_SIZE) {
			v = strings.TrimPrefix(v, CORE_DUMP_SIZE)
			r.CoreDumpSize, err = strconv.ParseUint(v, 10, 64)
			if err != nil {
				e = err
				return
			}
			continue
		}

		if strings.HasPrefix(v, MEMORY_USE) {
			v = strings.TrimPrefix(v, MEMORY_USE)
			r.MemoryUse, err = strconv.ParseUint(v, 10, 64)
			if err != nil {
				e = err
				return
			}
			continue
		}

		if strings.HasPrefix(v, MEMORY_LOCKED) {
			v = strings.TrimPrefix(v, MEMORY_LOCKED)
			r.MemoryLocked, err = strconv.ParseUint(v, 10, 64)
			if err != nil {
				e = err
				return
			}
			continue
		}

		if strings.HasPrefix(v, MAX_PROC) {
			v = strings.TrimPrefix(v, MAX_PROC)
			r.MaxProc, err = strconv.Atoi(v)
			if err != nil {
				e = err
				return
			}
			continue
		}

		if strings.HasPrefix(v, OPEN_FILES) {
			v = strings.TrimPrefix(v, OPEN_FILES)
			r.OpenFiles, err = strconv.Atoi(v)
			if err != nil {
				e = err
				return
			}
			continue
		}

		if strings.HasPrefix(v, V_MEMORY_USE) {
			v = strings.TrimPrefix(v, V_MEMORY_USE)
			r.VMemoryUse, err = strconv.ParseUint(v, 10, 64)
			if err != nil {
				e = err
				return
			}
			continue
		}

		if strings.HasPrefix(v, NTHR) {
			v = strings.TrimPrefix(v, NTHR)
			r.NThr, err = strconv.Atoi(v)
			if err != nil {
				e = err
				return
			}
			continue
		}

		if strings.HasPrefix(v, NSEMOP) {
			v = strings.TrimPrefix(v, NSEMOP)
			r.Nsemop, err = strconv.Atoi(v)
			if err != nil {
				e = err
				return
			}
			continue
		}

		if strings.HasPrefix(v, WALL_CLOCK) {
			v = strings.TrimPrefix(v, WALL_CLOCK)
			r.WallClock, err = strconv.ParseInt(v, 10, 64)
			if err != nil {
				e = err
				return
			}
			continue
		}

		if strings.HasPrefix(v, PCPU) {
			// This metrics has to be divided by the number of overall CPUs
			// In order to avoid returning a number greater than 100
			v = strings.TrimPrefix(v, PCPU)
			pcpu, err := strconv.Atoi(v)
			if err != nil {
				e = err
				return
			}

			info, err := FreeBSDOsInfo.GetCpuInfo()
			if err != nil {
				e = err
				return
			}

			if pcpu == 0 {
				r.PCpu = 0
			} else {
				r.PCpu = int(math.Floor((float64(pcpu) / float64(info.OverallCpus)) + 0.5)) // Round to the nearest integer
			}

			if r.PCpu > 100 {
				r.PCpu = 100
			}

			continue
		}

		if strings.HasPrefix(v, READ_BPS) {
			v = strings.TrimPrefix(v, READ_BPS)
			r.ReadBps, err = strconv.ParseUint(v, 10, 64)
			if err != nil {
				e = err
				return
			}
			continue
		}

		if strings.HasPrefix(v, WRITE_BPS) {
			v = strings.TrimPrefix(v, WRITE_BPS)
			r.WriteBps, err = strconv.ParseUint(v, 10, 64)
			if err != nil {
				e = err
				return
			}
			continue
		}

		if strings.HasPrefix(v, READ_IOPS) {
			v = strings.TrimPrefix(v, READ_IOPS)
			r.ReadIoPs, err = strconv.ParseUint(v, 10, 64)
			if err != nil {
				e = err
				return
			}
			continue
		}

		if strings.HasPrefix(v, WRITE_IOPS) {
			v = strings.TrimPrefix(v, WRITE_IOPS)
			r.WriteIoPs, err = strconv.ParseUint(v, 10, 64)
			if err != nil {
				e = err
				return
			}
			continue
		}
	}

	return
}
