package FreeBSDsysctls

import (
	"errors"
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

// Sysctl which returns a number of the available CPUs on a current system (sockets*cores*threads).
func SysctlHwNcpu() (int, error) {
	out, err := exec.Command("/sbin/sysctl", "-nq", "hw.ncpu").CombinedOutput()
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
	out, err := exec.Command("/sbin/sysctl", "-nq", "hw.vmm.maxcpu").CombinedOutput()
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
	out, err := exec.Command("/sbin/sysctl", "-nq", "hw.model").CombinedOutput()
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
	out, err := exec.Command("/sbin/sysctl", "-nq", "hw.machine").CombinedOutput()
	if err != nil {
		errorString := strings.TrimSpace(string(out)) + "; " + err.Error()
		return "", errors.New(errorString)
	}

	cpuArchitecture := strings.TrimSpace(string(out))
	return cpuArchitecture, nil
}

// Sysctl which returns a free memory pages, in pages.
// If you are looking to find a free memory in bytes, you'll need to multiply this value by the system's page size.
func SysctlVmStatsVmVfreecount() (uint64, error) {
	out, err := exec.Command("/sbin/sysctl", "-nq", "vm.stats.vm.v_free_count").CombinedOutput()
	if err != nil {
		errorString := strings.TrimSpace(string(out)) + "; " + err.Error()
		return 0, errors.New(errorString)
	}

	outValue := strings.TrimSpace(string(out))
	result, err := strconv.ParseUint(outValue, 10, 64)
	if err != nil {
		return 0, err
	}

	return result, nil
}

// Sysctl which returns a memory page size, in bytes.
func SysctlHwPagesize() (uint64, error) {
	out, err := exec.Command("/sbin/sysctl", "-nq", "hw.pagesize").CombinedOutput()
	if err != nil {
		errorString := strings.TrimSpace(string(out)) + "; " + err.Error()
		return 0, errors.New(errorString)
	}

	outValue := strings.TrimSpace(string(out))
	result, err := strconv.ParseUint(outValue, 10, 64)
	if err != nil {
		return 0, err
	}

	return result, nil
}

// Sysctl which returns an overall memory (RAM) size on any given system, in bytes.
func SysctlHwRealmem() (uint64, error) {
	out, err := exec.Command("/sbin/sysctl", "-nq", "hw.realmem").CombinedOutput()
	if err != nil {
		errorString := strings.TrimSpace(string(out)) + "; " + err.Error()
		return 0, errors.New(errorString)
	}

	outValue := strings.TrimSpace(string(out))
	result, err := strconv.ParseUint(outValue, 10, 64)
	if err != nil {
		return 0, err
	}

	return result, nil
}

// Sysctl which returns a system hostname as a string,
// or if there was an error -> "EMPTY_HOSTNAME" string
func SysctlKernHostname() (string, error) {
	result := ""
	emptyHostnameLabel := "EMPTY_HOSTNAME"

	out, err := exec.Command("/sbin/sysctl", "-nq", "kern.hostname").CombinedOutput()
	if err != nil {
		errorString := strings.TrimSpace(string(out)) + "; " + err.Error()
		return emptyHostnameLabel, errors.New(errorString)
	}

	result = strings.TrimSpace(string(out))
	if len(result) < 1 {
		result = emptyHostnameLabel
	}

	return result, nil
}

// Sysctl which returns a kernel boot time.
type BootTime struct {
	Sec   uint64
	USec  uint64
	Human string
}

func SysctlKernBoottime() (r BootTime, e error) {
	out, err := exec.Command("/sbin/sysctl", "-nq", "kern.boottime").CombinedOutput()
	if err != nil {
		e = fmt.Errorf("%s; %s", strings.TrimSpace(string(out)), err.Error())
		return
	}
	// Example output
	// { sec = 1704469832, usec = 847955 } Fri Jan  5 15:50:32 2024

	reMatchSec := regexp.MustCompile(`{\s+sec\s+=\s+\d+`)
	reMatchUSec := regexp.MustCompile(`usec\s+=\s+\d+`)
	reMatchNumber := regexp.MustCompile(`\d+`)
	reMatchCurly := regexp.MustCompile(`{.*?}`)
	outValue := strings.TrimSpace(string(out))

	secStr := reMatchSec.FindString(outValue)
	secStr = reMatchNumber.FindString(secStr)
	secInt, err := strconv.ParseUint(secStr, 10, 64)
	if err != nil {
		secInt = 0
	}
	r.Sec = secInt

	usecStr := reMatchUSec.FindString(outValue)
	usecStr = reMatchNumber.FindString(usecStr)
	usecInt, err := strconv.ParseUint(usecStr, 10, 64)
	if err != nil {
		usecInt = 0
	}
	r.USec = usecInt

	human := reMatchCurly.ReplaceAllString(outValue, "")
	human = strings.TrimSpace(human)
	r.Human = human

	return
}

// Sysctl which returns a size of the OpenZFS Arc.
func SysctlKstatZfsMiscArcstatsSize() (r uint64, e error) {
	out, err := exec.Command("/sbin/sysctl", "-nq", "kstat.zfs.misc.arcstats.size").CombinedOutput()
	if err != nil {
		e = fmt.Errorf("%s; %s", strings.TrimSpace(string(out)), err.Error())
		return
	}

	outValue := strings.TrimSpace(string(out))
	r, err = strconv.ParseUint(outValue, 10, 64)
	if err != nil {
		e = err
		return
	}

	return
}
