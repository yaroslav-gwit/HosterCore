// Copyright 2024 Hoster Authors. All rights reserved.
// Use of this source code is governed by an Apache License 2.0
// license that can be found in the LICENSE file.

package HosterVm

import (
	FreeBSDsysctls "HosterCore/internal/pkg/freebsd/sysctls"
	HosterHost "HosterCore/internal/pkg/hoster/host"
	HosterHostUtils "HosterCore/internal/pkg/hoster/host/utils"
	HosterNetwork "HosterCore/internal/pkg/hoster/network"
	HosterVmUtils "HosterCore/internal/pkg/hoster/vm/utils"
	"bufio"
	"errors"
	"os"
	"os/exec"
	"strings"
	"text/template"
	"time"

	"github.com/google/uuid"
)

type VmDeployInput struct {
	StartWhenRdy    bool   `json:"start_when_ready"`
	VCpus           int    `json:"vcpus"`
	VmName          string `json:"vm_name"`
	RAM             string `json:"ram"`
	NetworkName     string `json:"network_name"`
	IpAddress       string `json:"ip_address"`
	CustomDnsServer string `json:"custom_dns_server"`
	OsType          string `json:"os_type"`
	TargetDataset   string `json:"target_dataset"`
}

// Deploy a new VM. Returns an error if something went wrong.
func Deploy(input VmDeployInput) error {
	var err error

	// If the logger was already set, ignore this
	if !log.ConfigSet {
		log.SetFileLocation(HosterVmUtils.VM_AUDIT_LOG_LOCATION)
	}

	err = HosterVmUtils.ValidateResName(input.VmName)
	if err != nil {
		return err
	}

	// Initialize values
	c := ConfigOutput{}
	// Set CPU cores and RAM
	c.Cpus = input.VCpus
	c.Ram = input.RAM

	// Generate and set the root and gwitsuper users password
	c.RootPassword = HosterHostUtils.GenerateRandomPassword(33, true, true)
	c.GwitsuperPassword = HosterHostUtils.GenerateRandomPassword(33, true, true)
	// Generate and set CI instance ID
	c.InstanceId = HosterHostUtils.GenerateRandomPassword(5, false, true)

	// Generate correct VM name
	if len(input.VmName) < 1 || input.VmName == "test-vm" {
		c.VmName, err = HosterVmUtils.GenerateTestVmName(input.VmName)
		if err != nil {
			return errors.New("could not generate vm name: " + err.Error())
		}
	} else {
		c.VmName = input.VmName
	}

	// emojlog.PrintLogMessage("Deploying new VM: "+c.VmName, emojlog.Info)

	// Generate and set random MAC address
	c.MacAddress, err = HosterVmUtils.GenerateMacAddress()
	if err != nil {
		return errors.New("could not generate vm name: " + err.Error())
	}

	if len(input.IpAddress) > 1 {
		c.IpAddress = input.IpAddress
	} else {
		// Generate and set random IP address (which is free in the pool of addresses)
		c.IpAddress, err = HosterHostUtils.GenerateNewRandomIp(input.NetworkName)
		if err != nil {
			return errors.New("could not generate the IP: " + err.Error())
		}
	}

	networkInfo, err := HosterNetwork.GetNetworkConfig()
	if err != nil {
		return errors.New("could not read the network config")
	}
	if len(input.NetworkName) < 1 {
		c.NetworkName = networkInfo[0].NetworkName
		c.Subnet = networkInfo[0].Subnet
		c.NakedSubnet = strings.Split(networkInfo[0].Subnet, "/")[1]
		c.Gateway = networkInfo[0].Gateway
		c.NetworkComment = networkInfo[0].Comment
	} else {
		for _, v := range networkInfo {
			if input.NetworkName == v.NetworkName {
				c.NetworkName = v.NetworkName
				c.Subnet = v.Subnet
				c.NakedSubnet = strings.Split(v.Subnet, "/")[1]
				c.Gateway = v.Gateway
				c.NetworkComment = v.Comment
			}
		}
		if len(c.NetworkName) < 1 {
			return errors.New("network name supplied doesn't exist")
		}
	}

	if len(input.CustomDnsServer) > 1 {
		c.DnsServer = input.CustomDnsServer
	} else {
		c.DnsServer = c.Gateway
	}

	if strings.Contains(c.VmName, "test") {
		c.Production = false
	} else {
		c.Production = true
	}

	// emojlog.PrintLogMessage("OS type used: "+ , emojlog.Debug)
	c.OsType = input.OsType
	c.OsComment = HosterVmUtils.GenerateOsComment(c.OsType)

	c.ParentHost, _ = FreeBSDsysctls.SysctlKernHostname()
	c.VncPassword = HosterHostUtils.GenerateRandomPassword(8, true, true)
	c.VncPort, err = HosterVmUtils.GenerateVncPort()
	if err != nil {
		return errors.New("could not generate vnc port: " + err.Error())
	}

	c.SshKeys, err = getSystemSshKeys()
	if err != nil {
		return errors.New("could not get ssh keys: " + err.Error())
	}

	// Generate template ciUserDataTemplate
	tmpl, err := template.New("ciUserDataTemplate").Parse(HosterVmUtils.CiUserDataTemplate)
	if err != nil {
		return errors.New("could not generate ciUserDataTemplate: " + err.Error())
	}

	var ciUserData strings.Builder
	if err := tmpl.Execute(&ciUserData, c); err != nil {
		return errors.New("could not generate ciUserDataTemplate: " + err.Error())
	}
	// fmt.Println(ciUserData.String())

	// Generate template ciNetworkConfigTemplate
	tmpl, err = template.New("ciNetworkConfigTemplate").Parse(HosterVmUtils.CiNetworkConfigTemplate)
	if err != nil {
		return errors.New("could not generate ciNetworkConfigTemplate: " + err.Error())
	}

	var ciNetworkConfig strings.Builder
	if err := tmpl.Execute(&ciNetworkConfig, c); err != nil {
		return errors.New("could not generate ciNetworkConfigTemplate: " + err.Error())
	}
	// fmt.Println(ciNetworkConfig.String())

	// Generate template ciMetaDataTemplate
	tmpl, err = template.New("ciMetaDataTemplate").Parse(HosterVmUtils.CiMetaDataTemplate)
	if err != nil {
		return errors.New("could not generate ciMetaDataTemplate: " + err.Error())
	}

	var ciMetaData strings.Builder
	if err := tmpl.Execute(&ciMetaData, c); err != nil {
		return errors.New("could not generate ciMetaDataTemplate: " + err.Error())
	}
	// fmt.Println(ciMetaData.String())

	zfsCloneResult, err := zfsDatasetClone(input.TargetDataset, input.OsType, c.VmName)
	if err != nil || !zfsCloneResult {
		return err
	}

	// Write config files
	// emojlog.PrintLogMessage("Writing CloudInit config files", emojlog.Debug)
	newVmFolder := "/" + input.TargetDataset + "/" + c.VmName
	vmConfigFileLocation := newVmFolder + "/" + HosterVmUtils.VM_CONFIG_NAME
	vmConfig := HosterVmUtils.VmConfig{}
	networkConfig := HosterVmUtils.VmNetwork{}
	diskConfig := HosterVmUtils.VmDisk{}
	diskConfigList := []HosterVmUtils.VmDisk{}
	vmConfig.CPUSockets = 1
	vmConfig.CPUCores = c.Cpus
	vmConfig.Memory = c.Ram
	vmConfig.Loader = "uefi"
	vmConfig.Production = c.Production
	vmConfig.OsType = c.OsType
	vmConfig.OsComment = c.OsComment
	vmConfig.Owner = "system"
	vmConfig.ParentHost = c.ParentHost

	networkConfig.NetworkAdaptorType = "virtio-net"
	networkConfig.NetworkBridge = c.NetworkName
	networkConfig.NetworkMac = c.MacAddress
	networkConfig.IPAddress = c.IpAddress
	networkConfig.Comment = c.NetworkComment
	vmConfig.Networks = append(vmConfig.Networks, networkConfig)

	diskConfig.DiskType = "nvme"
	diskConfig.DiskLocation = "internal"
	diskConfig.DiskImage = "disk0.img"
	diskConfig.Comment = "OS Disk"
	diskConfigList = append(diskConfigList, diskConfig)
	diskConfig.DiskType = "ahci-cd"
	diskConfig.DiskLocation = "internal"
	diskConfig.DiskImage = "seed.iso"
	diskConfig.Comment = "CloudInit ISO"
	diskConfigList = append(diskConfigList, diskConfig)
	vmConfig.Disks = append(vmConfig.Disks, diskConfigList...)

	vmConfig.IncludeHostSSHKeys = true
	vmConfig.VmSshKeys = c.SshKeys
	vmConfig.VncPort = c.VncPort
	vmConfig.VncPassword = c.VncPassword
	vmConfig.UUID = uuid.New().String()
	vmConfig.Description = "-"

	err = HosterVmUtils.ConfigFileWriter(vmConfig, vmConfigFileLocation)
	if err != nil {
		return err
	}

	// Create cloud init folder
	if _, err := os.Stat(newVmFolder + "/cloud-init-files"); os.IsNotExist(err) {
		err = os.Mkdir(newVmFolder+"/cloud-init-files", 0750)
		if err != nil {
			return err
		}
	}
	// Open /cloud-init-files/user-data for writing
	ciUserDataFileLocation, err := os.Create(newVmFolder + "/cloud-init-files/user-data")
	if err != nil {
		return errors.New(err.Error())
	}
	defer ciUserDataFileLocation.Close()
	// Create a new writer
	writer := bufio.NewWriter(ciUserDataFileLocation)
	// Write a string to the file
	str := ciUserData.String()
	_, err = writer.WriteString(str)
	if err != nil {
		return errors.New(err.Error())
	}
	// Flush the writer to ensure all data has been written to the file
	err = writer.Flush()
	if err != nil {
		return errors.New(err.Error())
	}

	// Open /cloud-init-files/network-config for writing
	ciNetworkFileLocation, err := os.Create(newVmFolder + "/cloud-init-files/network-config")
	if err != nil {
		return errors.New(err.Error())
	}
	defer ciNetworkFileLocation.Close()
	// Create a new writer
	writer = bufio.NewWriter(ciNetworkFileLocation)
	// Write a string to the file
	str = ciNetworkConfig.String()
	_, err = writer.WriteString(str)
	if err != nil {
		return errors.New(err.Error())
	}
	// Flush the writer to ensure all data has been written to the file
	err = writer.Flush()
	if err != nil {
		return errors.New(err.Error())
	}

	// Open /cloud-init-files/meta-data for writing
	ciMetaDataFileLocation, err := os.Create(newVmFolder + "/cloud-init-files/meta-data")
	if err != nil {
		return errors.New(err.Error())
	}
	defer ciMetaDataFileLocation.Close()
	// Create a new writer
	writer = bufio.NewWriter(ciMetaDataFileLocation)
	// Write a string to the file
	str = ciMetaData.String()
	_, err = writer.WriteString(str)
	if err != nil {
		return errors.New(err.Error())
	}
	// Flush the writer to ensure all data has been written to the file
	err = writer.Flush()
	if err != nil {
		return errors.New(err.Error())
	}

	err = createCiIso(newVmFolder)
	if err != nil {
		return errors.New(err.Error())
	}

	err = HosterHostUtils.ReloadDns()
	if err != nil {
		return err
	}

	// Start the VM when all of the above is complete
	if input.StartWhenRdy {
		time.Sleep(time.Second * 1)
		// err := VmStart(c.VmName, false, false, false)
		err := Start(input.VmName, false, false)
		if err != nil {
			return err
		}
	}

	return nil
}

type ConfigOutput struct {
	Cpus              int
	Ram               string
	SshKeys           []HosterVmUtils.VmSshKey
	RootPassword      string
	GwitsuperPassword string
	InstanceId        string
	VmName            string
	NetworkName       string
	NetworkComment    string
	MacAddress        string
	IpAddress         string
	Subnet            string
	NakedSubnet       string
	Gateway           string
	DnsServer         string
	Production        bool
	OsType            string
	OsComment         string
	ParentHost        string
	VncPort           int
	VncPassword       string
}

func getSystemSshKeys() (r []HosterVmUtils.VmSshKey, e error) {
	conf, err := HosterHost.GetHostConfig()
	if err != nil {
		e = err
		return
	}

	for _, v := range conf.HostSSHKeys {
		r = append(r, HosterVmUtils.VmSshKey{KeyValue: v.KeyValue, Comment: v.Comment, KeyOwner: "System"})
	}

	return
}

func zfsDatasetClone(dsParent string, osType string, newVmName string) (bool, error) {
	vmTemplateExist := "/" + dsParent + "/template-" + osType + "/disk0.img"
	_, err := os.Stat(vmTemplateExist)

	if os.IsNotExist(err) {
		return false, errors.New("template dataset/disk image " + vmTemplateExist + " does not exist")
	} else if err != nil {
		return false, errors.New("error checking folder: " + err.Error())
	}

	vmTemplate := dsParent + "/template-" + osType

	snapName := "@deployment_" + newVmName + "_" + HosterHostUtils.GenerateRandomPassword(8, false, true)
	out, err := exec.Command("zfs", "snapshot", vmTemplate+snapName).CombinedOutput()
	if err != nil {
		return false, errors.New("could not execute zfs snapshot: " + string(out) + "; " + err.Error())
	}

	out, err = exec.Command("zfs", "clone", vmTemplate+snapName, dsParent+"/"+newVmName).CombinedOutput()
	if err != nil {
		return false, errors.New("could not execute zfs clone: " + string(out) + "; " + err.Error())
	}
	return true, nil
}

func createCiIso(vmFolder string) error {
	ciFolder := vmFolder + "/cloud-init-files/"
	out, err := exec.Command("genisoimage", "-output", vmFolder+"/seed.iso", "-volid", "cidata", "-joliet", "-rock", ciFolder+"user-data", ciFolder+"meta-data", ciFolder+"network-config").CombinedOutput()
	if err != nil {
		return errors.New("there was a problem generating an ISO: " + string(out) + "; " + err.Error())
	}

	// emojlog.PrintLogMessage("New CloudInit ISO has been created", emojlog.Info)
	return nil
}
