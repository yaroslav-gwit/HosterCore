package cmd

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/aquasecurity/table"
	"github.com/spf13/cobra"
)

var (
	jsonHostInfoOutput       bool
	jsonPrettyHostInfoOutput bool

	hostCmd = &cobra.Command{
		Use:   "host",
		Short: "Host related operations",
		Long:  `Host related operations.`,
		Run: func(cmd *cobra.Command, args []string) {
			hostMain()
			_ = getCpuInfo()
		},
	}
)

func hostMain() {
	if jsonPrettyHostInfoOutput {
		jsonOutputVar := jsonOutputHostInfo()
		jsonData, err := json.MarshalIndent(jsonOutputVar, "", "   ")
		if err != nil {
			log.Fatal("Function error: HostMain:", err)
		}
		fmt.Println(string(jsonData))
	} else if jsonHostInfoOutput {
		jsonOutputVar := jsonOutputHostInfo()
		jsonData, err := json.Marshal(jsonOutputVar)
		if err != nil {
			log.Fatal("Function error: HostMain:", err)
		}
		fmt.Println(string(jsonData))
	} else {
		var tHostname string
		var tLiveVms string
		var tSystemUptime string
		var tSystemRam = ramResponse{}
		var tSwapInfo swapInfoStruct
		var tArcSize string
		var tZrootInfo zrootInfoStruct
		var tPc2VcRatio string

		var wg = &sync.WaitGroup{}
		var err error
		wg.Add(1)
		go func() { defer wg.Done(); tHostname = GetHostName() }()
		wg.Add(1)
		go func() { defer wg.Done(); tLiveVms = getNumberOfRunningVms() }()
		wg.Add(1)
		go func() { defer wg.Done(); tPc2VcRatio, _ = getPc2VcRatio() }()
		wg.Add(1)
		go func() { defer wg.Done(); tSystemUptime = getSystemUptime() }()
		wg.Add(1)
		go func() { defer wg.Done(); tSystemRam = getHostRam() }()
		wg.Add(1)
		go func() { defer wg.Done(); tArcSize = getArcSize() }()
		wg.Add(1)
		go func() { defer wg.Done(); tZrootInfo = getZrootInfo() }()
		wg.Add(1)
		go func() {
			defer wg.Done()
			tSwapInfo, err = getSwapInfo()
			if err != nil {
				log.Fatal(err)
			}
		}()
		wg.Wait()

		t := table.New(os.Stdout)
		t.SetLineStyle(table.StyleBrightCyan)
		t.SetDividers(table.UnicodeRoundedDividers)
		t.SetHeaderStyle(table.StyleBold)

		t.SetAlignment(
			table.AlignLeft,   // Hostname
			table.AlignCenter, // Live VMs
			table.AlignCenter, // vCPU:pCPU
			table.AlignCenter, // System Uptime
			table.AlignCenter, // RAM
			table.AlignCenter, // SWAP
			table.AlignCenter, // ARC Size
			table.AlignCenter, // Zroot space
			table.AlignCenter, // Zpool status
		)

		t.SetHeaders("Brief Host Overview")
		t.SetHeaderColSpans(0, 9)
		t.AddHeaders(
			"Hostname",
			"Live VMs",
			"vCPU:pCPU Ratio",
			"System Uptime",
			"RAM (Used/Total)",
			"SWAP (Used/Total)",
			"ZFS ARC Size",
			"Zroot (Used/Total)",
			"Zroot Status",
		)

		t.AddRow(tHostname,
			tLiveVms,
			tPc2VcRatio,
			tSystemUptime,
			tSystemRam.used+"/"+tSystemRam.all,
			tSwapInfo.used+"/"+tSwapInfo.total,
			tArcSize,
			tZrootInfo.used+"/"+tZrootInfo.total,
			tZrootInfo.status,
		)

		t.Render()
	}
}

type jsonOutputHostInfoStruct struct {
	Hostname     string `json:"hostname"`
	LiveVms      int    `json:"live_vms"`
	AllVms		 int    `json:"all_vms"`
	BackupVms	 int    `json:"backup_vms"`
	SystemUptime string `json:"system_uptime"`
	CpuModel     string `json:"cpu_model"`
	CpuArchitecture     string `json:"cpu_architecture"`
	CpuSockets   int    `json:"cpu_sockets"`
	CpuCores     int    `json:"cpu_cores"`
	CpuThreads   int    `json:"cpu_threads"`
	RamTotal     string `json:"ram_total"`
	RamFree      string `json:"ram_free"`
	RamUsed      string `json:"ram_used"`
	SwapTotal    string `json:"swap_total"`
	SwapUsed     string `json:"swap_used"`
	SwapFree     string `json:"swap_free"`
	ArcSize      string `json:"zfs_acr_size"`
	ZrootTotal   string `json:"zroot_total"`
	ZrootUsed    string `json:"zroot_used"`
	ZrootFree    string `json:"zroot_free"`
	ZrootStatus  string `json:"zroot_status"`
}

func jsonOutputHostInfo() jsonOutputHostInfoStruct {
	var tHostname string
	var tLiveVms int
	var tAllVms int
	var tBackupVms int
	var tSystemUptime string
	var tSystemRam = ramResponse{}
	var tSwapInfo swapInfoStruct
	var tCpuInfo CpuInfo
	var tArcSize string
	var tZrootInfo zrootInfoStruct

	var wg = &sync.WaitGroup{}
	var err error
	wg.Add(1)
	go func() { defer wg.Done(); tHostname = GetHostName() }()
	wg.Add(1)
	go func() { defer wg.Done(); tAllVms, tLiveVms, tBackupVms = vmNumbersOverview() }()
	wg.Add(1)
	go func() { defer wg.Done(); tSystemUptime = getSystemUptime() }()
	wg.Add(1)
	go func() { defer wg.Done(); tSystemRam = getHostRam() }()
	wg.Add(1)
	go func() { defer wg.Done(); tArcSize = getArcSize() }()
	wg.Add(1)
	go func() { defer wg.Done(); tZrootInfo = getZrootInfo() }()
	wg.Add(1)
	go func() { defer wg.Done(); tCpuInfo = getCpuInfo() }()

	wg.Add(1)
	go func() {
		defer wg.Done()
		tSwapInfo, err = getSwapInfo()
		if err != nil {
			log.Fatal(err)
		}
	}()
	wg.Wait()

	jsonOutputVar := jsonOutputHostInfoStruct{}
	jsonOutputVar.Hostname = tHostname
	jsonOutputVar.LiveVms = tLiveVms
	jsonOutputVar.AllVms = tAllVms
	jsonOutputVar.BackupVms = tBackupVms
	jsonOutputVar.SystemUptime = tSystemUptime
	jsonOutputVar.RamTotal = tSystemRam.all
	jsonOutputVar.RamFree = tSystemRam.free
	jsonOutputVar.RamUsed = tSystemRam.used
	jsonOutputVar.SwapTotal = tSwapInfo.total
	jsonOutputVar.SwapUsed = tSwapInfo.used
	jsonOutputVar.SwapFree = tSwapInfo.free
	jsonOutputVar.ArcSize = tArcSize
	jsonOutputVar.ZrootTotal = tZrootInfo.total
	jsonOutputVar.ZrootUsed = tZrootInfo.used
	jsonOutputVar.ZrootFree = tZrootInfo.free
	jsonOutputVar.ZrootStatus = tZrootInfo.status
	jsonOutputVar.CpuModel = tCpuInfo.Model
	jsonOutputVar.CpuArchitecture = tCpuInfo.Architecture
	jsonOutputVar.CpuSockets = tCpuInfo.Sockets
	jsonOutputVar.CpuCores = tCpuInfo.Cores
	jsonOutputVar.CpuThreads = tCpuInfo.Threads

	return jsonOutputVar
}

type ramResponse struct {
	free string
	used string
	all  string
}

// type swapResponse struct {
// 	free string
// 	all  string
// }

// type zrootUsageResponse struct {
// 	free string
// 	all  string
// }

func getHostRam() ramResponse {
	// GET SYSCTL "vm.stats.vm.v_free_count" AND RETURN THE VALUE
	var vFreeCount string
	vFreeCountArg1 := "sysctl"
	vFreeCountArg2 := "-nq"
	vFreeCountArg3 := "vm.stats.vm.v_free_count"

	cmd := exec.Command(vFreeCountArg1, vFreeCountArg2, vFreeCountArg3)
	stdout, err := cmd.Output()
	if err != nil {
		fmt.Println("Func freeRam/vFreeCount: There has been an error:", err)
		os.Exit(1)
	} else {
		vFreeCount = string(stdout)
	}

	var vFreeCountList []string
	for _, i := range strings.Split(vFreeCount, "\n") {
		if len(i) > 1 {
			vFreeCountList = append(vFreeCountList, i)
		}
	}
	vFreeCount = vFreeCountList[0]

	var realMem string
	realMemArg1 := "sysctl"
	realMemArg2 := "-nq"
	realMemArg3 := "hw.realmem"

	cmd = exec.Command(realMemArg1, realMemArg2, realMemArg3)
	stdout, err = cmd.Output()
	if err != nil {
		fmt.Println("Func freeRam/vFreeCount: There has been an error:", err)
		os.Exit(1)
	} else {
		realMem = string(stdout)
	}

	var realMemList []string
	for _, i := range strings.Split(realMem, "\n") {
		if len(i) > 1 {
			realMemList = append(realMemList, i)
		}
	}
	realMem = realMemList[0]

	// GET SYSCTL "hw.pagesize" AND RETURN THE VALUE
	var hwPagesize string
	var hwPagesizeArg1 = "sysctl"
	var hwPagesizeArg2 = "-nq"
	var hwPagesizeArg3 = "hw.pagesize"
	cmd = exec.Command(hwPagesizeArg1, hwPagesizeArg2, hwPagesizeArg3)
	stdout, err = cmd.Output()
	if err != nil {
		fmt.Println("Func freeRam/hwPagesize: There has been an error:", err)
		os.Exit(1)
	} else {
		hwPagesize = string(stdout)
	}
	var hwPagesizeList []string
	for _, i := range strings.Split(hwPagesize, "\n") {
		if len(i) > 1 {
			hwPagesizeList = append(hwPagesizeList, i)
		}
	}
	hwPagesize = hwPagesizeList[0]

	vFreeCountInt, _ := strconv.Atoi(vFreeCount)
	hwPagesizeInt, _ := strconv.Atoi(hwPagesize)
	realMemInt, _ := strconv.Atoi(realMem)

	finalResultFree := vFreeCountInt * hwPagesizeInt
	finalResultReal := realMemInt
	finalResultUsed := (finalResultReal - finalResultFree)

	ramResponseVar := ramResponse{}
	ramResponseVar.free = ByteConversion(finalResultFree)
	ramResponseVar.used = ByteConversion(finalResultUsed)
	ramResponseVar.all = ByteConversion(finalResultReal)

	return ramResponseVar
}

func getArcSize() string {
	// GET SYSCTL "vm.stats.vm.v_free_count" AND RETURN THE VALUE
	var arcSize string
	arcSizeArg1 := "sysctl"
	arcSizeArg2 := "-nq"
	arcSizeArg3 := "kstat.zfs.misc.arcstats.size"

	cmd := exec.Command(arcSizeArg1, arcSizeArg2, arcSizeArg3)
	stdout, err := cmd.Output()
	if err != nil {
		fmt.Println("Func getArcSize/arcSize: There has been an error:", err)
		os.Exit(1)
	} else {
		arcSize = string(stdout)
	}

	var arcSizeList []string
	for _, i := range strings.Split(arcSize, "\n") {
		if len(i) > 1 {
			arcSizeList = append(arcSizeList, i)
		}
	}

	acrSizeInt, _ := strconv.Atoi(arcSizeList[0])
	return ByteConversion(acrSizeInt)
}

func getNumberOfRunningVms() string {
	files, err := os.ReadDir("/dev/vmm/")
	var finalResult string
	if err != nil {
		// fmt.Println("funcError getNumberOfRunningVms: " + err.Error())
		// os.Exit(1)
		finalResult = "0"
	} else {
		finalResult = strconv.Itoa(len(files))

	}

	return finalResult
}

// func getZrootStatus() string {
// 	var zrootStatus string
// 	var zrootStatusArg1 = "zpool"
// 	var zrootStatusArg2 = "status"
// 	var zrootStatusArg3 = "zroot"

// 	var cmd = exec.Command(zrootStatusArg1, zrootStatusArg2, zrootStatusArg3)
// 	var stdout, err = cmd.Output()
// 	if err != nil {
// 		fmt.Println("Func getZrootStatus/zrootStatus: There has been an error:", err)
// 		os.Exit(1)
// 	} else {
// 		zrootStatus = string(stdout)
// 	}
// 	var zrootStatusList []string
// 	for _, i := range strings.Split(zrootStatus, "\n") {
// 		if len(i) > 1 {
// 			zrootStatusList = append(zrootStatusList, i)
// 		}
// 	}

// 	var r, _ = regexp.Compile(".*state:.*")
// 	for _, i := range zrootStatusList {
// 		var reMatch = r.MatchString(i)
// 		if reMatch {
// 			zrootStatus = i
// 		}
// 	}

// 	zrootStatus = strings.Replace(zrootStatus, "state:", "", -1)
// 	zrootStatus = strings.Replace(zrootStatus, " ", "", -1)
// 	if zrootStatus == "ONLINE" {
// 		zrootStatus = "Online"
// 	} else {
// 		zrootStatus = "Problem!"
// 	}

// 	var finalResult = zrootStatus

// 	return finalResult
// }

// func getFreeZfsSpace() string {
// 	var zrootFree string
// 	var zrootFreeArg1 = "zfs"
// 	var zrootFreeArg2 = "list"
// 	var zrootFreeArg3 = "zroot"

// 	var cmd = exec.Command(zrootFreeArg1, zrootFreeArg2, zrootFreeArg3)
// 	var stdout, err = cmd.Output()
// 	if err != nil {
// 		fmt.Println("Func getFreeZfsSpace/zrootFree: There has been an error:", err)
// 		os.Exit(1)
// 	} else {
// 		zrootFree = string(stdout)
// 	}
// 	var zrootFreeList []string
// 	for _, i := range strings.Split(zrootFree, " ") {
// 		if len(i) > 1 {
// 			zrootFreeList = append(zrootFreeList, i)
// 		}
// 	}

// 	zrootFree = zrootFreeList[6]
// 	zrootFree = strings.Replace(zrootFree, " ", "", -1)

// 	var finalResult = zrootFree
// 	return finalResult
// }

type swapInfoStruct struct {
	total string
	used  string
	free  string
}

func getSwapInfo() (swapInfoStruct, error) {
	swapInfoVar := swapInfoStruct{}
	stdout, stderr := exec.Command("swapinfo").Output()
	if stderr != nil {
		return swapInfoStruct{}, stderr
	}

	var swapInfoString string
	for _, v := range strings.Split(string(stdout), "\n") {
		if len(v) > 0 {
			swapInfoString = v
		}
	}

	var swapInfoList []string
	reSplitSpace := regexp.MustCompile(`\s+`)
	for _, v := range reSplitSpace.Split(swapInfoString, -1) {
		if len(v) > 0 {
			swapInfoList = append(swapInfoList, v)
		}
	}

	swapTotalBytes, _ := strconv.Atoi(swapInfoList[1])
	swapTotalBytes = swapTotalBytes * 1024
	swapUsedBytes, _ := strconv.Atoi(swapInfoList[2])
	swapUsedBytes = swapUsedBytes * 1024
	swapFreeBytes, _ := strconv.Atoi(swapInfoList[3])
	swapFreeBytes = swapFreeBytes * 1024

	swapInfoVar.total = ByteConversion(swapTotalBytes)
	swapInfoVar.free = ByteConversion(swapFreeBytes)
	swapInfoVar.used = ByteConversion(swapUsedBytes)

	return swapInfoVar, nil
}

func getSystemUptime() string {
	var systemUptime string

	stdout, err := exec.Command("sysctl", "-nq", "kern.boottime").Output()
	if err != nil {
		fmt.Println("Func getSystemUptime/systemUptime: There has been an error:", err)
		os.Exit(1)
	} else {
		systemUptime = string(stdout)
	}

	var systemUptimeList []string
	for _, i := range strings.Split(systemUptime, " ") {
		if len(i) > 1 {
			systemUptimeList = append(systemUptimeList, i)
		}
	}

	systemUptime = systemUptimeList[1]
	systemUptime = strings.Replace(systemUptime, ",", "", -1)
	systemUptime = strings.Replace(systemUptime, " ", "", -1)

	systemUptimeInt, _ := strconv.ParseInt(systemUptime, 10, 64)
	unixTime := time.Unix(systemUptimeInt, 0)

	timeSince := time.Since(unixTime).Seconds()
	secondsModulus := int(timeSince) % 60.0

	minutesSince := (timeSince - float64(secondsModulus)) / 60.0
	minutesModulus := int(minutesSince) % 60.0

	hoursSince := (minutesSince - float64(minutesModulus)) / 60
	hoursModulus := int(hoursSince) % 24

	daysSince := (int(hoursSince) - hoursModulus) / 24

	result := strconv.Itoa(daysSince) + "d "
	result = result + strconv.Itoa(hoursModulus) + "h "
	result = result + strconv.Itoa(minutesModulus) + "m "
	result = result + strconv.Itoa(secondsModulus) + "s"

	return result
}

type zrootInfoStruct struct {
	total  string
	used   string
	free   string
	status string
}

func getZrootInfo() zrootInfoStruct {
	zrootInfoVar := zrootInfoStruct{}
	stdout, err := exec.Command("zpool", "list", "-p", "zroot").Output()
	if err != nil {
		log.Fatal("Could not run zpool list")
	}

	reSplitSpace := regexp.MustCompile(`\s+`)
	var zrootInfoList []string
	zrootInfoList = append(zrootInfoList, reSplitSpace.Split(string(stdout), -1)...)
	// for _, v := range reSplitSpace.Split(string(stdout), -1) {
	// 	zrootInfoList = append(zrootInfoList, v)
	// }

	if zrootInfoList[20] == "ONLINE" {
		zrootInfoVar.status = "Healthy"
	}

	zrootInfoTotalInt, _ := strconv.ParseInt(zrootInfoList[12], 10, 64)
	zrootInfoUsedInt, _ := strconv.ParseInt(zrootInfoList[13], 10, 64)
	zrootInfoFreeInt, _ := strconv.ParseInt(zrootInfoList[14], 10, 64)

	zrootInfoVar.total = ByteConversion(int(zrootInfoTotalInt))
	zrootInfoVar.used = ByteConversion(int(zrootInfoUsedInt))
	zrootInfoVar.free = ByteConversion(int(zrootInfoFreeInt))

	return zrootInfoVar
}

func GetHostName() string {
	// GET SYSCTL "vm.stats.vm.v_free_count" AND RETURN THE VALUE
	stdout, err := exec.Command("sysctl", "-nq", "kern.hostname").Output()
	if err != nil {
		fmt.Println("Func GetHostName/hostName: There has been an error:", err)
		os.Exit(1)
	}
	var hostNameList []string
	for _, i := range strings.Split(string(stdout), "\n") {
		if len(i) > 1 {
			hostNameList = append(hostNameList, i)
		}
	}
	return hostNameList[0]
}

type CpuInfo struct {
	Model 		   string
	Architecture   string
	Sockets  	   int
	Cores 		   int
	Threads        int
	OverallCpus    int
}

func getCpuInfo() CpuInfo {
	result := CpuInfo{}

	command, err := exec.Command("sysctl", "-nq", "hw.model").CombinedOutput()
	if err != nil {
		fmt.Println("Error", err.Error())
	}
	cpuModel := strings.TrimSpace(string(command))
	result.Model = cpuModel

	command, err = exec.Command("sysctl", "-nq", "hw.machine").CombinedOutput()
	if err != nil {
		fmt.Println("Error", err.Error())
	}
	cpuArch := strings.TrimSpace(string(command))
	result.Architecture = cpuArch

	command, err = exec.Command("sysctl", "-nq", "hw.ncpu").CombinedOutput()
	if err != nil {
		fmt.Println("Error", err.Error())
	}
	cpuCores := strings.TrimSpace(string(command))
	result.OverallCpus, _ = strconv.Atoi(cpuCores)

	command, err = exec.Command("grep", "-i", "threads", "/var/run/dmesg.boot").CombinedOutput()
	if err != nil {
		fmt.Println("Error", err.Error())
	}
	dmesg := string(command)
	reCpuInfoMatch := regexp.MustCompile(`.*package.*core.*thread`)
	reStripPrefix := regexp.MustCompile(`.*FreeBSD/SMP:\s`)
	for _, v := range strings.Split(dmesg, "\n") {
		if reCpuInfoMatch.MatchString(v) {
			dmesg = reStripPrefix.ReplaceAllString(v, "")
			break
		}
	}
	var tempCpuInfoList []string
	var tempCpuInfoListStripped []string
	
	// The loop was replaced with the one liner right below it
	// for _, v := range strings.Split(dmesg, " x ") {
		// 	tempCpuInfoList = append(tempCpuInfoList, v)
	// }
	tempCpuInfoList = append(tempCpuInfoList, strings.Split(dmesg, " x ")...)

	for _, v := range tempCpuInfoList {
		v := strings.Split(v, " ")[0]
		tempCpuInfoListStripped = append(tempCpuInfoListStripped, v)
	}

	result.Sockets, _ = strconv.Atoi(tempCpuInfoListStripped[0])
	result.Cores, _ = strconv.Atoi(tempCpuInfoListStripped[1])
	result.Threads, _ = strconv.Atoi(tempCpuInfoListStripped[2])

	return result
}

func getPc2VcRatio() (string, float64) {
	command, err := exec.Command("sysctl", "-nq", "hw.ncpu").CombinedOutput()
	if err != nil {
		fmt.Println("Error", err.Error())
	}
	
	cpuLogicalCores, err := strconv.Atoi(strings.TrimSpace(string(command)))
	if err != nil {
		cpuLogicalCores = 1
		fmt.Println(err.Error())
	}

	coresUsed := 0
	for _, v := range getAllVms() {
		temp, _ := getVmInfo(v)
		if vmLiveCheck(v) {
			coresUsed = coresUsed + (temp.CpuCores * temp.CpuSockets)
		}
	}

	result := float64(coresUsed / cpuLogicalCores)
	var ratio string
	if result < 1 {
		ratio = LIGHT_GREEN + "<1" + NC
	} else if result >= 1 && result <= 3 {
		ratio = LIGHT_GREEN + fmt.Sprintf("%.0f", result) + ":1" + NC
	} else if result >= 3 && result <= 5 {
		ratio = LIGHT_YELLOW + fmt.Sprintf("%.0f", result) + ":1" + NC
	} else if result > 5 {
		ratio = LIGHT_RED + fmt.Sprintf("%.0f", result) + ":1" + NC
	}

	return ratio, result
}

// Returns 3 integers: All VMs, Online/Live VMs, and Backup VMs
func vmNumbersOverview() (int, int, int) {
	allVms := 0
	onlineVms := 0
	backupVms := 0
	for _, v := range getAllVms() {
		allVms = allVms + 1
		if vmLiveCheck(v) {
			onlineVms = onlineVms + 1
		}
		tempConf := vmConfig(v)
		if tempConf.ParentHost != GetHostName() {
			backupVms = backupVms + 1
		}
	}
	return allVms, onlineVms, backupVms
}

func ByteConversion(bytes int) string {
	// SET TO KB
	var firstIteration = bytes / 1024
	var iterationTitle = "K"

	// SET TO MB
	if firstIteration > 1024 {
		firstIteration = firstIteration / 1024
		iterationTitle = "M"
	}

	// SET TO GB
	var firstIterationFloat = 0.0
	if firstIteration > 1024 {
		firstIterationFloat = float64(firstIteration) / 1024.0
		iterationTitle = "G"
	}

	// FORMAT THE OUTPUT
	var finalResult string
	if firstIterationFloat > 0.0 {
		finalResult = fmt.Sprintf("%.2f", firstIterationFloat) + iterationTitle
	} else {
		finalResult = strconv.Itoa(firstIteration) + iterationTitle
	}
	return finalResult
}


// Console color outputs
const LIGHT_RED = "\033[38;5;203m"
const LIGHT_GREEN = "\033[0;92m"
const LIGHT_YELLOW = "\033[0;93m"
const NC = "\033[0m"
