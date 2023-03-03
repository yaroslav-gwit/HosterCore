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
		Use:   "info [vm name]",
		Short: "Print out the VM Info",
		Long:  `Print out the VM Info.`,
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			printVmInfo(args[0])
		},
	}
)

func printVmInfo(vmName string) {
	vmInfo, err := getVmInfo(vmName)
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

type vmInfoStruct struct {
	VmName             string `json:"vm_name,omitempty"`
	MainIpAddress      string `json:"main_ip_address,omitempty"`
	VmStatusLive       bool   `json:"vm_status_live"`
	VmStatusEncrypted  bool   `json:"vm_status_encrypted"`
	VmStatusProduction bool   `json:"vm_status_production"`
	VmStatusBackup     bool   `json:"vm_status_backup"`
	CpuSockets         int    `json:"cpu_sockets,omitempty"`
	CpuCores           int    `json:"cpu_cores,omitempty"`
	RamAmount          string `json:"ram_amount,omitempty"`
	VncPort            int    `json:"vnc_port,omitempty"`
	VncPassword        string `json:"vnc_password,omitempty"`
	OsType             string `json:"os_type,omitempty"`
	OsComment          string `json:"os_comment,omitempty"`
	VmUptime           string `json:"vm_uptime,omitempty"`
	VmDescription      string `json:"vm_description,omitempty"`
	ParentHost         string `json:"parent_host,omitempty"`
	Uptime             string `json:"uptime,omitempty"`
	OsDiskTotal        string `json:"os_disk_total,omitempty"`
	OsDiskUsed         string `json:"os_disk_used,omitempty"`
}

func getVmInfo(vmName string) (vmInfoStruct, error) {
	var vmInfoVar = vmInfoStruct{}

	allVms := getAllVms()
	if slices.Contains(allVms, vmName) {
		_ = true
	} else {
		return vmInfoStruct{}, errors.New("VM is not found in the system")
	}

	vmInfoVar.VmName = vmName
	vmInfoVar.ParentHost = GetHostName()

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

		if vmConfigVar.ParentHost == vmInfoVar.ParentHost {
			vmInfoVar.VmStatusBackup = false
		} else {
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
	go func() { defer wg.Done(); vmInfoVar.VmStatusLive = vmLiveCheck(vmName) }()

	wg.Add(1)
	go func() { defer wg.Done(); vmInfoVar.OsDiskTotal = getOsDiskFullSize(vmName) }()

	wg.Add(1)
	go func() { defer wg.Done(); vmInfoVar.OsDiskUsed = getOsDiskUsed(vmName) }()

	wg.Add(1)
	go func() { defer wg.Done(); vmInfoVar.Uptime = getVmUptimeNew(vmName) }()

	wg.Wait()
	return vmInfoVar, nil
}
