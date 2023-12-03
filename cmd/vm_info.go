package cmd

import (
	"encoding/json"
	"errors"
	"log"
	"strconv"

	"github.com/spf13/cobra"
	"golang.org/x/exp/slices"
)

var (
	jsonVmInfo       bool
	jsonPrettyVmInfo bool

	vmInfoCmd = &cobra.Command{
		Use:   "info [vmName]",
		Short: "Print out the VM Info",
		Long:  `Print out the VM Info.`,
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			checkInitFile()
			printVmInfo(args[0])
		},
	}
)

func printVmInfo(vmName string) {
	vmInfo, err := GetVmInfo(vmName, false)
	if err != nil {
		log.Fatal(err)
	}
	if jsonPrettyVmInfo {
		jsonPretty, err := json.MarshalIndent(vmInfo, "", "   ")
		if err != nil {
			log.Fatal(err)
		}
		println(string(jsonPretty))
	} else {
		jsonOutput, err := json.Marshal(vmInfo)
		if err != nil {
			log.Fatal(err)
		}
		println(string(jsonOutput))
	}
}

type VmInfoStruct struct {
	VmName             string `json:"vm_name"`
	MainIpAddress      string `json:"main_ip_address"`
	VmStatusLive       bool   `json:"vm_status_live"`
	VmStatusEncrypted  bool   `json:"vm_status_encrypted"`
	VmStatusProduction bool   `json:"vm_status_production"`
	VmStatusBackup     bool   `json:"vm_status_backup"`
	CpuSockets         int    `json:"cpu_sockets"`
	CpuCores           int    `json:"cpu_cores"`
	RamAmount          string `json:"ram_amount"`
	VncPort            int    `json:"vnc_port"`
	VncPassword        string `json:"vnc_password"`
	OsType             string `json:"os_type"`
	OsComment          string `json:"os_comment"`
	VmUptime           string `json:"vm_uptime"`
	VmDescription      string `json:"vm_description"`
	ParentHost         string `json:"parent_host"`
	CurrentHost        string `json:"current_host"`
	Uptime             string `json:"uptime"`
	OsDiskTotal        string `json:"os_disk_total"`
	OsDiskUsed         string `json:"os_disk_used"`
}

func GetVmInfo(vmName string, useGlobalUptime bool) (VmInfoStruct, error) {
	var vmInfoVar = VmInfoStruct{}

	allVms := getAllVms()
	if slices.Contains(allVms, vmName) {
		_ = true
	} else {
		return VmInfoStruct{}, errors.New("VM " + vmName + " is not found on this system")
	}

	vmInfoVar.VmName = vmName
	vmInfoVar.CurrentHost = GetHostName()

	wg.Add(1)
	go func() {
		defer wg.Done()
		vmConfigVar := vmConfig(vmName)
		vmInfoVar.MainIpAddress = vmConfigVar.Networks[0].IPAddress

		if vmConfigVar.LiveStatus == "production" || vmConfigVar.LiveStatus == "prod" {
			vmInfoVar.VmStatusProduction = true
		} else {
			vmInfoVar.VmStatusProduction = false
		}

		if vmInfoVar.CurrentHost == vmConfigVar.ParentHost {
			vmInfoVar.ParentHost = vmInfoVar.CurrentHost
			vmInfoVar.VmStatusBackup = false
		} else {
			vmInfoVar.ParentHost = vmConfigVar.ParentHost
			vmInfoVar.VmStatusBackup = true
		}

		cpuSockets, err := strconv.Atoi(vmConfigVar.CPUSockets)
		if err != nil {
			log.Fatal(err)
		}
		vmInfoVar.CpuSockets = cpuSockets

		cpuCores, err := strconv.Atoi(vmConfigVar.CPUCores)
		if err != nil {
			log.Fatal(err)
		}
		vmInfoVar.CpuCores = cpuCores
		vmInfoVar.RamAmount = vmConfigVar.Memory
		vncPort, err := strconv.Atoi(vmConfigVar.VncPort)
		if err != nil {
			log.Fatal(err)
		}
		vmInfoVar.VncPort = vncPort
		vmInfoVar.VncPassword = vmConfigVar.VncPassword
		vmInfoVar.OsType = vmConfigVar.OsType
		vmInfoVar.OsComment = vmConfigVar.OsComment
		vmInfoVar.VmDescription = vmConfigVar.Description
	}()

	wg.Add(1)
	go func() { defer wg.Done(); vmInfoVar.VmStatusEncrypted = encryptionCheck(vmName) }()

	wg.Add(1)
	go func() { defer wg.Done(); vmInfoVar.VmStatusLive = VmLiveCheck(vmName) }()

	wg.Add(1)
	go func() { defer wg.Done(); vmInfoVar.OsDiskTotal = getOsDiskFullSize(vmName) }()

	wg.Add(1)
	go func() { defer wg.Done(); vmInfoVar.OsDiskUsed = getOsDiskUsed(vmName) }()

	wg.Add(1)
	go func() { defer wg.Done(); vmInfoVar.Uptime = getVmUptimeNew(vmName, useGlobalUptime) }()

	wg.Wait()
	return vmInfoVar, nil
}
