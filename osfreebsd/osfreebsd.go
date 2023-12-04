package osfreebsd

import (
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

// Sets all the CPU info and returns it as a pointer to it's struct
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

// Sysctl which returns a number of CPUs on a current system (sockets X cores X threads)
func SysctlHwNcpu() (int, error) {
	out, err := exec.Command("sysctl", "-nq", "hw.ncpu").CombinedOutput()
	if err != nil {
		errorString := strings.TrimSpace(string(out)) + "; " + err.Error()
		return 0, errors.New(errorString)
	}

	allCpus := strings.TrimSpace(string(out))
	allCpusNum, err := strconv.Atoi(allCpus)
	if err != nil {
		return 0, err
	}

	return allCpusNum, nil
}

// Sysctl which returns a maximum number of CPUs that can be used by the Bhyve on a single VM
func SysctlHwVmmMaxcpu() (int, error) {
	out, err := exec.Command("sysctl", "-nq", "hw.vmm.maxcpu").CombinedOutput()
	if err != nil {
		errorString := strings.TrimSpace(string(out)) + "; " + err.Error()
		return 0, errors.New(errorString)
	}

	allCpus := strings.TrimSpace(string(out))
	allCpusNum, err := strconv.Atoi(allCpus)
	if err != nil {
		return 0, err
	}

	return allCpusNum, nil
}

// Sysctl which returns a CPU model, and strips some of the ambiguous symbols
func SysctlHwModel() (string, error) {
	out, err := exec.Command("sysctl", "-nq", "hw.model").CombinedOutput()
	if err != nil {
		errorString := strings.TrimSpace(string(out)) + "; " + err.Error()
		return "", errors.New(errorString)
	}

	cpuModel := strings.TrimSpace(string(out))
	reStripCpuModel := regexp.MustCompile(`\(R\)|\(TM\)|@\s|CPU\s`)
	cpuModel = reStripCpuModel.ReplaceAllString(cpuModel, "")

	return cpuModel, nil
}

// Sysctl which returns a CPU architecture, for example amd64, arm64, etc
func SysctlHwMachine() (string, error) {
	out, err := exec.Command("sysctl", "-nq", "hw.machine").CombinedOutput()
	if err != nil {
		errorString := strings.TrimSpace(string(out)) + "; " + err.Error()
		return "", errors.New(errorString)
	}

	cpuArchitecture := strings.TrimSpace(string(out))
	return cpuArchitecture, nil
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
