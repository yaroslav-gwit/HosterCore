package FreeBSDOsInfo

import (
	FreeBSDsysctls "HosterCore/internal/pkg/freebsd/sysctls"
	"errors"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
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
		c.Threads = 1
		// return c, err
	}

	allCpus, err := FreeBSDsysctls.SysctlHwNcpu()
	if err != nil {
		return c, err
	}
	c.Cores = allCpus / (c.Threads * c.Sockets)

	c.Model, err = FreeBSDsysctls.SysctlHwModel()
	if err != nil {
		return c, err
	}

	c.Architecture, err = FreeBSDsysctls.SysctlHwMachine()
	if err != nil {
		return c, err
	}

	c.OverallCpus, err = FreeBSDsysctls.SysctlHwNcpu()
	if err != nil {
		return c, err
	}

	return c, nil
}

// Returns a slice of strings from `dmesg.boot` split at a carriage return
func DmesgCpuGrep() ([]string, error) {
	out, err := exec.Command("grep", "-i", "package", "/var/run/dmesg.boot").CombinedOutput()
	if err != nil {
		errorString := strings.TrimSpace(string(out)) + "; " + err.Error()
		return []string{}, errors.New(errorString)
	}

	r := []string{}
	for _, v := range strings.Split(string(out), "\n") {
		v = strings.TrimSpace(v)
		if len(v) > 0 {
			r = append(r, v)
		}
	}

	return r, nil
}
