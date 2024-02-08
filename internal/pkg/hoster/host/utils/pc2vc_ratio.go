package HosterHostUtils

import (
	FreeBSDsysctls "HosterCore/internal/pkg/freebsd/sysctls"
	termcolors "HosterCore/internal/pkg/hoster/terminal_colours"
	"fmt"
)

// If this ratio is > 6:1 you should really stop deploying more VMs to avoid performance issues
//
// This is a good read to better understand the Pc2Vc value:
// https://download3.vmware.com/vcat/vmw-vcloud-architecture-toolkit-spv1-webworks/index.html#page/Core%20Platform/Architecting%20a%20vSphere%20Compute%20Platform/Architecting%20a%20vSphere%20Compute%20Platform.1.019.html
// func GetPc2VcRatio() (string, float64) {
// 	command, err := exec.Command("sysctl", "-nq", "hw.ncpu").CombinedOutput()
// 	if err != nil {
// 		fmt.Println("Error", err.Error())
// 	}

// 	cpuLogicalCores, err := strconv.Atoi(strings.TrimSpace(string(command)))
// 	if err != nil {
// 		cpuLogicalCores = 1
// 		fmt.Println(err.Error())
// 	}

// 	coresUsed := 0
// 	for _, v := range getAllVms() {
// 		temp, _ := GetVmInfo(v, true)
// 		if VmLiveCheck(v) {
// 			coresUsed = coresUsed + (temp.CpuCores * temp.CpuSockets)
// 		}
// 	}

// 	result := float64(coresUsed / cpuLogicalCores)
// 	var ratio string
// 	if result < 1 {
// 		ratio = termcolors.LIGHT_GREEN + "<1" + termcolors.NC
// 	} else if result >= 1 && result <= 3 {
// 		ratio = termcolors.LIGHT_GREEN + fmt.Sprintf("%.0f", result) + ":1" + termcolors.NC
// 	} else if result >= 3 && result <= 5 {
// 		ratio = termcolors.LIGHT_YELLOW + fmt.Sprintf("%.0f", result) + ":1" + termcolors.NC
// 	} else if result > 5 {
// 		ratio = termcolors.LIGHT_RED + fmt.Sprintf("%.0f", result) + ":1" + termcolors.NC
// 	}

// 	return ratio, result
// }

// Same as the above, but doesn't need to iterate over all VMs every time, because you can submit a number of used CPUs.
func GetPc2VcRatioLazy(cpusUsed int) (string, float64) {
	cpusAvailable, err := FreeBSDsysctls.SysctlHwNcpu()
	if err != nil {
		cpusAvailable = 0
	}

	result := float64(cpusUsed / cpusAvailable)
	var ratio string
	if result < 1 {
		ratio = termcolors.LIGHT_GREEN + "<1" + termcolors.NC
	} else if result >= 1 && result <= 3 {
		ratio = termcolors.LIGHT_GREEN + fmt.Sprintf("%.0f", result) + ":1" + termcolors.NC
	} else if result >= 3 && result <= 5 {
		ratio = termcolors.LIGHT_YELLOW + fmt.Sprintf("%.0f", result) + ":1" + termcolors.NC
	} else if result > 5 {
		ratio = termcolors.LIGHT_RED + fmt.Sprintf("%.0f", result) + ":1" + termcolors.NC
	}

	return ratio, result
}
