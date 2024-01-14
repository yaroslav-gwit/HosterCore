package FreeBSDsysctls

import (
	"errors"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

// Sysctl which returns a number of CPUs on a current system (sockets X cores X threads)
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
