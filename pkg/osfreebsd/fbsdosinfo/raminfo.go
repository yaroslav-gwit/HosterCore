package fbsdosinfo

import (
	"HosterCore/pkg/byteconversion"
	"HosterCore/pkg/osfreebsd/fbsdsysctls"
)

type RamInfo struct {
	RamFreeHuman    string
	RamFreeBytes    uint64
	RamUsedHuman    string
	RamUsedBytes    uint64
	RamOverallHuman string
	RamOverallBytes uint64
}

// Returns a structured RAM information for your FreeBSD system
func GetRamInfo() (RamInfo, error) {
	r := RamInfo{}

	hwPagesize, err := fbsdsysctls.SysctlHwPagesize()
	if err != nil {
		return r, err
	}
	freePages, err := fbsdsysctls.SysctlVmStatsVmVfreecount()
	if err != nil {
		return r, err
	}
	realMem, err := fbsdsysctls.SysctlHwRealmem()
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
