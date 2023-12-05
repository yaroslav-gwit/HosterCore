package osfreebsd

import (
	"errors"
	"regexp"
	"strconv"
)

type CpuInfo struct {
	Model        string
	Architecture string
	Sockets      int
	Cores        int
	Threads      int
	OverallCpus  int
}

// Returns a structured CPU information for your FreeBSD system
func GetCpuInfo() (CpuInfo, error) {
	c := CpuInfo{}
	reMatchNumber := regexp.MustCompile(`\d+`)

	dmesg, err := DmesgCpuGrep()
	if err != nil {
		return c, err
	}

	reMatchSockets := regexp.MustCompile(`\s+(\d+)\s+package`)
	socketsString := reMatchSockets.FindString(dmesg[0])
	socketsString = reMatchNumber.FindString(socketsString)
	c.Sockets, err = strconv.Atoi(socketsString)
	if err != nil {
		return c, errors.New("socket err " + err.Error())
	}

	reMatchThreads := regexp.MustCompile(`x\s+(\d+)\s+(?:hardware\s+)?threads`)
	threadsString := reMatchThreads.FindString(dmesg[0])
	threadsString = reMatchNumber.FindString(threadsString)
	c.Threads, err = strconv.Atoi(threadsString)
	if err != nil {
		return c, err
	}

	allCpus, err := SysctlHwNcpu()
	if err != nil {
		return c, err
	}
	c.Cores = allCpus / (c.Threads * c.Sockets)

	c.Model, err = SysctlHwModel()
	if err != nil {
		return c, err
	}

	c.Architecture, err = SysctlHwMachine()
	if err != nil {
		return c, err
	}

	c.OverallCpus, err = SysctlHwNcpu()
	if err != nil {
		return c, err
	}

	return c, nil
}

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

	hwPagesize, err := SysctlHwPagesize()
	if err != nil {
		return r, err
	}
	freePages, err := SysctlVmStatsVmVfreecount()
	if err != nil {
		return r, err
	}
	realMem, err := SysctlHwRealmem()
	if err != nil {
		return r, err
	}

	resultFreeBytes := freePages * hwPagesize
	resultUsedBytes := realMem - resultFreeBytes

	r.RamFreeHuman = BytesToHuman(resultFreeBytes)
	r.RamFreeBytes = resultFreeBytes

	r.RamUsedHuman = BytesToHuman(resultUsedBytes)
	r.RamUsedBytes = resultUsedBytes

	r.RamOverallHuman = BytesToHuman(realMem)
	r.RamOverallBytes = realMem

	return r, nil
}
