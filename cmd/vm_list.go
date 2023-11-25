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

	"facette.io/natsort"
	"github.com/aquasecurity/table"
	"github.com/spf13/cobra"
)

var (
	jsonOutputVm       bool
	jsonPrettyOutputVm bool
	tableUnixOutputVm  bool

	vmListCmd = &cobra.Command{
		Use:   "list",
		Short: "VM list",
		Long:  `VM list in the form of tables, json, or json pretty`,
		Run: func(cmd *cobra.Command, args []string) {
			err := checkInitFile()
			if err != nil {
				log.Fatal(err.Error())
			}

			vmListMain()
		},
	}
)

var allVmsUptime []string
var allVmsUptimeTimesUsed int
var wg = &sync.WaitGroup{}

func vmListMain() {
	if jsonOutputVm {
		vmList := vmJsonOutput()
		jsonPrint, err := json.Marshal(vmList)
		if err != nil {
			log.Fatal(err)
		}
		println(string(jsonPrint))
	} else if jsonPrettyOutputVm {
		vmList := vmJsonOutput()
		jsonPrint, err := json.MarshalIndent(vmList, "", "   ")
		if err != nil {
			log.Fatal(err)
		}
		println(string(jsonPrint))
	} else {
		vmTableOutput()
	}
}

func vmJsonOutput() []string {
	return getAllVms()
}

func vmTableOutput() {
	wg.Add(2)
	var vmInfo []string
	var thisHostName string
	go func() { defer wg.Done(); vmInfo = getAllVms() }()
	go func() { defer wg.Done(); thisHostName = GetHostName() }()
	wg.Wait()

	var ID = 0
	var vmLive string
	var vmEncrypted string
	var vmProduction string
	var vmConfigVar VmConfigStruct

	var t = table.New(os.Stdout)
	t.SetAlignment(table.AlignRight, //ID
		table.AlignLeft,   // VM Name
		table.AlignCenter, // VM Status
		table.AlignCenter, // CPU Sockets
		table.AlignCenter, // CPU Cores
		table.AlignCenter, // RAM
		table.AlignLeft,   // Main IP
		table.AlignLeft,   // OS Comment
		table.AlignLeft,   // VM Uptime
		table.AlignCenter, // OS Disk Used
		table.AlignLeft)   // Description

	if tableUnixOutputVm {
		t.SetDividers(table.Dividers{
			ALL: " ",
			NES: " ",
			NSW: " ",
			NEW: " ",
			ESW: " ",
			NE:  " ",
			NW:  " ",
			SW:  " ",
			ES:  " ",
			EW:  " ",
			NS:  " ",
		})
		t.SetRowLines(false)
		t.SetBorderTop(false)
		t.SetBorderBottom(false)
	} else {
		t.SetHeaders("Hoster VMs")
		t.SetHeaderColSpans(0, 11)

		t.AddHeaders(
			"#",
			"VM\nName",
			"VM\nStatus",
			"CPU\nSockets",
			"CPU\nCores",
			"VM\nMemory",
			"Main IP\nAddress",
			"OS\nType",
			"VM\nUptime",
			"OS Disk\n(Used/Total)",
			"VM\nDescription")

		t.SetLineStyle(table.StyleBrightCyan)
		t.SetDividers(table.UnicodeRoundedDividers)
		t.SetHeaderStyle(table.StyleBold)
	}

	for _, vmName := range vmInfo {
		var vmOsDiskFullSize string
		var vmOsDiskFree string
		var vmUptimeVar string

		vmConfigVar = vmConfig(vmName)
		ID = ID + 1
		vmLive = vmLiveCheckString(vmName)

		wg.Add(4)
		go func() { defer wg.Done(); vmOsDiskFullSize = getOsDiskFullSize(vmName) }()
		go func() { defer wg.Done(); vmOsDiskFree = getOsDiskUsed(vmName) }()
		go func() { defer wg.Done(); vmEncrypted = encryptionCheckString(vmName) }()
		go func() { defer wg.Done(); vmUptimeVar = getVmUptimeNew(vmName, true) }()

		if vmConfigVar.ParentHost != thisHostName {
			wg.Add(1)
			go func() {
				vmLive = "💾"
				vmSnaps, err := GetSnapshotInfo(vmName, true)
				if err != nil {
					vmConfigVar.Description = "💾⏩ " + vmConfigVar.ParentHost
					wg.Done()
					return
				}
				lastSnap := ""
				if len(vmSnaps) == 1 {
					lastSnap = vmSnaps[0].Name
				} else if len(vmSnaps) > 1 {
					lastSnap = vmSnaps[len(vmSnaps)-1].Name
				} else {
					vmConfigVar.Description = "💾⏩ " + vmConfigVar.ParentHost
					wg.Done()
					return
				}
				lastSnap = strings.Split(lastSnap, "@")[1]
				vmConfigVar.Description = "💾⏩ " + vmConfigVar.ParentHost + " 🕔 " + lastSnap
				wg.Done()
			}()
		}

		wg.Wait()

		if VmIsInProduction(vmConfigVar.LiveStatus) {
			vmProduction = "🔁"
		} else {
			vmProduction = ""
		}

		t.AddRow(strconv.Itoa(ID),
			vmName,
			vmLive+vmEncrypted+vmProduction,
			vmConfigVar.CPUSockets,
			vmConfigVar.CPUCores,
			vmConfigVar.Memory,
			vmConfigVar.Networks[0].IPAddress,
			vmConfigVar.OsComment,
			vmUptimeVar,
			vmOsDiskFree+"/"+vmOsDiskFullSize,
			vmConfigVar.Description)
	}

	t.Render()
}

// Monkey patch to export getAllVms() func
func GetAllVms() []string {
	return getAllVms()
}

// Return a simple list of all VMs on this system
func getAllVms() []string {
	var zfsDatasets []string
	var configFileName = "/vm_config.json"
	hostConfig, err := GetHostConfig()
	if err != nil {
		fmt.Println(err)
	}
	if len(hostConfig.ActiveDatasets) < 1 {
		zfsDatasets = append(zfsDatasets, "zroot/vm-encrypted")
		zfsDatasets = append(zfsDatasets, "zroot/vm-unencrypted")
	} else {
		zfsDatasets = append(zfsDatasets, hostConfig.ActiveDatasets...)
	}

	var vmListSorted []string
	for _, dataset := range zfsDatasets {
		var dsFolder = "/" + dataset + "/"
		var _, err = os.Stat(dsFolder)
		if !os.IsNotExist(err) {
			vmFolders, err := os.ReadDir(dsFolder)

			if err != nil {
				fmt.Println("Error!", err)
				os.Exit(1)
			}

			for _, vmFolder := range vmFolders {
				info, _ := os.Stat(dsFolder + vmFolder.Name())
				if info.IsDir() {
					var _, err = os.Stat(dsFolder + vmFolder.Name() + configFileName)
					if !os.IsNotExist(err) {
						vmListSorted = append(vmListSorted, vmFolder.Name())
					}
				}
			}
		}
	}

	natsort.Sort(vmListSorted)
	return vmListSorted
}

func encryptionCheck(vmName string) bool {
	var zfsDatasets []string
	var dsFolder string
	var finalResponse bool
	zfsDatasets = append(zfsDatasets, "zroot/vm-encrypted")
	// zfsDatasets = append(zfsDatasets, "zroot/vm-unencrypted")

	for _, dataset := range zfsDatasets {
		dsFolder = "/" + dataset + "/"
		var _, err = os.Stat(dsFolder + vmName)
		if !os.IsNotExist(err) {
			finalResponse = true
			if finalResponse {
				break
			}
		} else {
			finalResponse = false
		}
	}
	return finalResponse
}

func encryptionCheckString(vmName string) string {
	var result = encryptionCheck(vmName)
	if result {
		return "🔒"
	} else {
		return ""
	}
}

func VmLiveCheck(vmName string) bool {
	var _, err = os.Stat("/dev/vmm/" + vmName)
	if !os.IsNotExist(err) {
		return true
	} else {
		return false
	}
}

func vmLiveCheckString(vmName string) string {
	var vmLive = VmLiveCheck(vmName)
	if vmLive {
		return "🟢"
	} else {
		return "🔴"
	}
}

func VmConfig(vmName string) VmConfigStruct {
	return vmConfig(vmName)
}

// Returns active VM config, using VmConfigStruct struct
func vmConfig(vmName string) VmConfigStruct {
	var configFile = getVmFolder(vmName) + "/vm_config.json"
	var jsonData = VmConfigStruct{}
	var content, err = os.ReadFile(configFile)
	if err != nil {
		fmt.Println("vmConfig Function Error: ", err)
	}
	if err := json.Unmarshal(content, &jsonData); err != nil {
		panic(err)
	}
	return jsonData
}

// Returns a folder containing VM files, using this format:
// "/zroot/vm-encrypted/vmName"
func getVmFolder(vmName string) string {
	hostConfig, err := GetHostConfig()
	if err != nil {
		fmt.Println(err)
	}

	var zfsDatasets []string
	var dsFolder string
	var finalResponse string

	if len(hostConfig.ActiveDatasets) < 1 {
		zfsDatasets = append(zfsDatasets, "zroot/vm-encrypted")
		zfsDatasets = append(zfsDatasets, "zroot/vm-unencrypted")
	} else {
		zfsDatasets = append(zfsDatasets, hostConfig.ActiveDatasets...)
	}

	for _, dataset := range zfsDatasets {
		dsFolder = "/" + dataset + "/"
		var _, err = os.Stat(dsFolder + vmName)
		if !os.IsNotExist(err) {
			finalResponse = dsFolder + vmName
		}
	}
	return finalResponse
}

func getVmUptimeNew(vmName string, useGlobal bool) string {
	var vmsUptime []string
	if len(allVmsUptime) > 0 && useGlobal && allVmsUptimeTimesUsed < 15 {
		allVmsUptimeTimesUsed += 1
		vmsUptime = allVmsUptime
	} else {
		if !useGlobal {
			allVmsUptimeTimesUsed = 0
		}

		var cmd = exec.Command("ps", "axwww", "-o", "etimes,command")
		stdout, err := cmd.Output()
		if err != nil {
			log.Fatal("getVmUptimeNew Error: ", err)
		}

		allVmsUptime = strings.Split(string(stdout), "\n")
		vmsUptime = allVmsUptime
	}

	rexMatchVmName, _ := regexp.Compile(`.*bhyve: ` + vmName + `.*`)
	var finalResult string
	for i, v := range vmsUptime {
		if rexMatchVmName.MatchString(v) {
			v = strings.TrimSpace(v)
			v = strings.Split(v, " ")[0]

			var vmUptimeInt, _ = strconv.ParseInt(v, 10, 64)
			var secondsModulus = int(vmUptimeInt) % 60.0

			var minutesSince = (float64(vmUptimeInt) - float64(secondsModulus)) / 60.0
			var minutesModulus = int(minutesSince) % 60.0

			var hoursSince = (minutesSince - float64(minutesModulus)) / 60
			var hoursModulus = int(hoursSince) % 24

			var daysSince = (int(hoursSince) - hoursModulus) / 24

			finalResult = strconv.Itoa(daysSince) + "d "
			finalResult = finalResult + strconv.Itoa(hoursModulus) + "h "
			finalResult = finalResult + strconv.Itoa(minutesModulus) + "m "
			finalResult = finalResult + strconv.Itoa(secondsModulus) + "s"

			break
		} else if i == (len(vmsUptime) - 1) {
			finalResult = "0s"
		}
	}
	return finalResult
}

func getOsDiskFullSize(vmName string) string {
	var filePath = getVmFolder(vmName) + "/disk0.img"
	var osDiskLs string
	var osDiskLsArg1 = "ls"
	var osDiskLsArg2 = "-ahl"
	var osDiskLsArg3 = filePath

	var cmd = exec.Command(osDiskLsArg1, osDiskLsArg2, osDiskLsArg3)
	var stdout, err = cmd.Output()
	if err != nil {
		fmt.Println("Func getOsDiskFullSize: There has been an error:", err)
		os.Exit(1)
	} else {
		osDiskLs = string(stdout)
	}

	var osDiskLsList []string
	for _, i := range strings.Split(osDiskLs, " ") {
		if len(i) > 1 {
			osDiskLsList = append(osDiskLsList, i)
			// fmt.Println(n, i)
		}
	}
	osDiskLs = osDiskLsList[3]
	return osDiskLs
}

func getOsDiskUsed(vmName string) string {
	var osDiskDu string
	var filePath = getVmFolder(vmName) + "/disk0.img"

	var cmd = exec.Command("du", "-h", filePath)
	var stdout, err = cmd.Output()
	if err != nil {
		fmt.Println("Func getOsDiskFullSize: There has been an error:", err)
		os.Exit(1)
	} else {
		osDiskDu = string(stdout)
	}

	var osDiskDuList []string
	for _, i := range strings.Split(osDiskDu, "/") {
		if len(i) > 1 {
			osDiskDuList = append(osDiskDuList, i)
		}
	}

	osDiskDu = strings.TrimSpace(osDiskDuList[0])
	return osDiskDu
}

type VmDiskStruct struct {
	DiskType     string `json:"disk_type"`
	DiskLocation string `json:"disk_location"`
	DiskImage    string `json:"disk_image"`
	Comment      string `json:"comment"`
}

type VmNetworkStruct struct {
	NetworkAdaptorType string `json:"network_adaptor_type"`
	NetworkBridge      string `json:"network_bridge"`
	NetworkMac         string `json:"network_mac"`
	IPAddress          string `json:"ip_address"`
	Comment            string `json:"comment"`
}

type VmSshKey struct {
	KeyOwner string `json:"key_owner"`
	KeyValue string `json:"key_value"`
	Comment  string `json:"comment"`
}

type VmConfigStruct struct {
	CPUSockets             string            `json:"cpu_sockets"`
	CPUCores               string            `json:"cpu_cores"`
	Memory                 string            `json:"memory"`
	Loader                 string            `json:"loader"`
	LiveStatus             string            `json:"live_status"`
	OsType                 string            `json:"os_type"`
	OsComment              string            `json:"os_comment"`
	Owner                  string            `json:"owner"`
	ParentHost             string            `json:"parent_host"`
	Networks               []VmNetworkStruct `json:"networks"`
	Disks                  []VmDiskStruct    `json:"disks"`
	IncludeHostwideSSHKeys bool              `json:"include_hostwide_ssh_keys"`
	VmSshKeys              []VmSshKey        `json:"vm_ssh_keys"`
	VncPort                string            `json:"vnc_port"`
	VncPassword            string            `json:"vnc_password"`
	Description            string            `json:"description"`
	UUID                   string            `json:"uuid,omitempty"`
	VGA                    string            `json:"vga,omitempty"`
	Passthru               []string          `json:"passthru,omitempty"`
	DisableXHCI            bool              `json:"disable_xhci,omitempty"`
	VncResolution          int               `json:"vnc_resolution,omitempty"`
}
