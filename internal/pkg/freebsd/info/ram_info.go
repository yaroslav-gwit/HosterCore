// Copyright 2023 Hoster Authors. All rights reserved.
// Use of this source code is governed by an Apache License 2.0
// license that can be found in the LICENSE file.

package FreeBSDOsInfo

import (
	"HosterCore/internal/pkg/byteconversion"
	FreeBSDsysctls "HosterCore/internal/pkg/freebsd/sysctls"
)

type RamInfo struct {
	RamFreeHuman    string `json:"ram_free_human"`
	RamFreeBytes    uint64 `json:"ram_free_bytes"`
	RamUsedHuman    string `json:"ram_used_human"`
	RamUsedBytes    uint64 `json:"ram_used_bytes"`
	RamOverallHuman string `json:"ram_overall_human"`
	RamOverallBytes uint64 `json:"ram_overall_bytes"`
}

// Returns a structured RAM information for your FreeBSD system
func GetRamInfo() (RamInfo, error) {
	r := RamInfo{}

	hwPagesize, err := FreeBSDsysctls.SysctlHwPagesize()
	if err != nil {
		return r, err
	}
	freePages, err := FreeBSDsysctls.SysctlVmStatsVmVfreecount()
	if err != nil {
		return r, err
	}
	realMem, err := FreeBSDsysctls.SysctlHwRealmem()
	if err != nil {
		return r, err
	}

	resultFreeBytes := freePages * hwPagesize
	resultUsedBytes := realMem - resultFreeBytes

	r.RamFreeHuman = byteconversion.BytesToHuman(resultFreeBytes)
	r.RamFreeBytes = resultFreeBytes

	r.RamUsedHuman = byteconversion.BytesToHuman(resultUsedBytes)
	r.RamUsedBytes = resultUsedBytes

	r.RamOverallHuman = byteconversion.BytesToHuman(realMem)
	r.RamOverallBytes = realMem

	return r, nil
}
