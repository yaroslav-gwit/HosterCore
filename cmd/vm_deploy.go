package cmd

import (
	"HosterCore/emojlog"
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"net"
	"os"
	"os/exec"
	"path"
	"regexp"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/spf13/cobra"
	"golang.org/x/exp/slices"
)

var (
	vmName                 string
	networkName            string
	deployIpAddress        string
	deployDnsServer        string
	osType                 string
	osTypeAlias            string
	zfsDataset             string
	vmDeployCpus           int
	vmDeployRam            string
	vmDeployStartWhenReady bool
	vmDeployFromIso        bool
	vmDeployIsoFilePath    string

	vmDeployCmd = &cobra.Command{
		Use:   "deploy",
		Short: "Deploy a new VM",
		Long:  `Deploy a new VM, using the pre-defined templates or ISO files`,
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			checkInitFile()

			if len(osTypeAlias) > 0 {
				osType = osTypeAlias
			}

			var err error
			if vmDeployFromIso {
				err = deployVmFromIso(vmName, networkName, osType, zfsDataset, vmDeployCpus, vmDeployRam, vmDeployStartWhenReady, vmDeployIsoFilePath)
				if err != nil {
					log.Fatal(err)
				}
			} else {
				err = deployVmMain(vmName, networkName, osType, zfsDataset, vmDeployCpus, vmDeployRam, vmDeployStartWhenReady)
				if err != nil {
					log.Fatal(err)
				}
			}
		},
	}
)

type ConfigOutputStruct struct {
	Cpus              string
	Ram               string
	SshKeys           []VmSshKey
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
	LiveStatus        string
	OsType            string
	OsComment         string
	ParentHost        string
	VncPort           string
	VncPassword       string
}

func deployVmMain(vmName string, networkName string, osType string, dsParent string, cpus int, ram string, startWhenReady bool) error {
	vmNameError := checkVmNameInput(vmName)
	if vmNameError != nil {
		return vmNameError
	}

	// Initialize values
	c := ConfigOutputStruct{}
	var err error

	// Set CPU cores and RAM
	c.Cpus = strconv.Itoa(cpus)
	c.Ram = ram

	// Generate and set the root and gwitsuper users password
	c.RootPassword = generateRandomPassword(33, true, true)
	if err != nil {
		return errors.New("could not generate random password for root user: " + err.Error())
	}
	c.GwitsuperPassword = generateRandomPassword(33, true, true)
	if err != nil {
		return errors.New("could not generate random password for gwitsuper user: " + err.Error())
	}

	// Generate and set CI instance ID
	c.InstanceId = generateRandomPassword(5, false, true)
	if err != nil {
		return errors.New("could not generate random instance id: " + err.Error())
	}

	// Generate correct VM name
	c.VmName, err = generateVmName(vmName)
	if err != nil {
		return errors.New("could not generate vm name: " + err.Error())
	}

	emojlog.PrintLogMessage("Deploying new VM: "+c.VmName, emojlog.Info)

	// Generate and set random MAC address
	c.MacAddress, err = generateRandomMacAddress()
	if err != nil {
		return errors.New("could not generate vm name: " + err.Error())
	}

	if len(deployIpAddress) > 1 {
		c.IpAddress = deployIpAddress
	} else {
		// Generate and set random IP address (which is free in the pool of addresses)
		c.IpAddress, err = generateNewIp(networkName)
		if err != nil {
			return errors.New("could not generate the IP: " + err.Error())
		}
	}

	networkInfo, err := networkInfo()
	if err != nil {
		return errors.New("could not read the network config")
	}
	if len(networkName) < 1 {
		c.NetworkName = networkInfo[0].Name
		c.Subnet = networkInfo[0].Subnet
		c.NakedSubnet = strings.Split(networkInfo[0].Subnet, "/")[1]
		c.Gateway = networkInfo[0].Gateway
		c.NetworkComment = networkInfo[0].Comment
	} else {
		for _, v := range networkInfo {
			if networkName == v.Name {
				c.NetworkName = v.Name
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

	if len(deployDnsServer) > 1 {
		c.DnsServer = deployDnsServer
	} else {
		c.DnsServer = c.Gateway
	}

	reMatchTest := regexp.MustCompile(`.*test`)
	if reMatchTest.MatchString(c.VmName) {
		c.LiveStatus = "testing"
	} else {
		c.LiveStatus = "production"
	}

	emojlog.PrintLogMessage("OS type used: "+osType, emojlog.Debug)
	c.OsType = osType
	switch c.OsType {
	case "debian11":
		c.OsComment = "Debian 11"
	case "debian12":
		c.OsComment = "Debian 12"
	case "ubuntu2004":
		c.OsComment = "Ubuntu 20.04"
	case "ubuntu2204":
		c.OsComment = "Ubuntu 22.04"
	case "almalinux8":
		c.OsComment = "AlmaLinux 8"
	case "rockylinux8":
		c.OsComment = "RockyLinux 8"
	case "freebsd13ufs":
		c.OsComment = "FreeBSD 13 UFS"
	case "freebsd13zfs":
		c.OsComment = "FreeBSD 13 ZFS"
	case "windows10":
		c.OsComment = "Windows 10"
	case "windows11":
		c.OsComment = "Windows 11"
	case "windows-srv19":
		c.OsComment = "Windows Server 19"
	case "windowssrv19":
		c.OsComment = "Windows Server 19"
	case "windows-srv22":
		c.OsComment = "Windows Server 22"
	case "windowssrv22":
		c.OsComment = "Windows Server 22"
	default:
		c.OsComment = "Custom OS"
	}

	c.ParentHost = GetHostName()
	c.VncPort = generateRandomVncPort()
	c.VncPassword = generateRandomPassword(8, true, true)
	if err != nil {
		return errors.New("could not generate vnc port: " + err.Error())
	}

	c.SshKeys, err = getSystemSshKeys()
	if err != nil {
		return errors.New("could not get ssh keys: " + err.Error())
	}

	// Generate template ciUserDataTemplate
	tmpl, err := template.New("ciUserDataTemplate").Parse(ciUserDataTemplate)
	if err != nil {
		return errors.New("could not generate ciUserDataTemplate: " + err.Error())
	}

	var ciUserData strings.Builder
	if err := tmpl.Execute(&ciUserData, c); err != nil {
		return errors.New("could not generate ciUserDataTemplate: " + err.Error())
	}
	// fmt.Println(ciUserData.String())

	// Generate template ciNetworkConfigTemplate
	tmpl, err = template.New("ciNetworkConfigTemplate").Parse(ciNetworkConfigTemplate)
	if err != nil {
		return errors.New("could not generate ciNetworkConfigTemplate: " + err.Error())
	}

	var ciNetworkConfig strings.Builder
	if err := tmpl.Execute(&ciNetworkConfig, c); err != nil {
		return errors.New("could not generate ciNetworkConfigTemplate: " + err.Error())
	}
	// fmt.Println(ciNetworkConfig.String())

	// Generate template ciMetaDataTemplate
	tmpl, err = template.New("ciMetaDataTemplate").Parse(ciMetaDataTemplate)
	if err != nil {
		return errors.New("could not generate ciMetaDataTemplate: " + err.Error())
	}

	var ciMetaData strings.Builder
	if err := tmpl.Execute(&ciMetaData, c); err != nil {
		return errors.New("could not generate ciMetaDataTemplate: " + err.Error())
	}
	// fmt.Println(ciMetaData.String())

	zfsCloneResult, err := zfsDatasetClone(dsParent, osType, c.VmName)
	if err != nil || !zfsCloneResult {
		return err
	}

	// Write config files
	emojlog.PrintLogMessage("Writing CloudInit config files", emojlog.Debug)
	newVmFolder := "/" + dsParent + "/" + c.VmName
	vmConfigFileLocation := newVmFolder + "/vm_config.json"
	vmConfig := VmConfigStruct{}
	networkConfig := VmNetworkStruct{}
	diskConfig := VmDiskStruct{}
	diskConfigList := []VmDiskStruct{}
	vmConfig.CPUSockets = "1"
	vmConfig.CPUCores = c.Cpus
	vmConfig.Memory = c.Ram
	vmConfig.Loader = "uefi"
	vmConfig.LiveStatus = c.LiveStatus
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

	vmConfig.IncludeHostwideSSHKeys = true
	vmConfig.VmSshKeys = c.SshKeys
	vmConfig.VncPort = c.VncPort
	vmConfig.VncPassword = c.VncPassword
	vmConfig.Description = "-"

	err = vmConfigFileWriter(vmConfig, vmConfigFileLocation)
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

	err = ReloadDnsServer()
	if err != nil {
		return err
	}

	// Start the VM when all of the above is complete
	if startWhenReady {
		time.Sleep(time.Second * 1)
		err := VmStart(c.VmName, false, false, false)
		if err != nil {
			return err
		}
	}

	return nil
}

func deployVmFromIso(vmName string, networkName string, osType string, dsParent string, cpus int, ram string, startWhenReady bool, isoPath string) error {
	if len(isoPath) < 1 {
		return errors.New("please, specify which ISO file will be used for the installation")
	}

	if !FileExists(isoPath) {
		return errors.New("the ISO file you've specified doesn't exist")
	}

	vmNameError := checkVmNameInput(vmName)
	if vmNameError != nil {
		return vmNameError
	}

	// Initialize values
	c := ConfigOutputStruct{}
	var err error

	// Set CPU cores and RAM
	c.Cpus = strconv.Itoa(cpus)
	c.Ram = ram

	// Generate and set the root and gwitsuper users password
	c.RootPassword = generateRandomPassword(33, true, true)
	c.GwitsuperPassword = generateRandomPassword(33, true, true)
	// Generate and set CI instance ID
	c.InstanceId = generateRandomPassword(5, false, true)

	// Generate correct VM name
	c.VmName, err = generateVmName(vmName)
	if err != nil {
		return errors.New("could not generate vm name: " + err.Error())
	}

	emojlog.PrintLogMessage("Deploying new VM: "+c.VmName, emojlog.Info)

	// Generate and set random MAC address
	c.MacAddress, err = generateRandomMacAddress()
	if err != nil {
		return errors.New("could not generate vm name: " + err.Error())
	}

	if len(deployIpAddress) > 1 {
		c.IpAddress = deployIpAddress
	} else {
		// Generate and set random IP address (which is free in the pool of addresses)
		c.IpAddress, err = generateNewIp(networkName)
		if err != nil {
			return errors.New("could not generate the IP: " + err.Error())
		}
	}

	networkInfo, err := networkInfo()
	if err != nil {
		return errors.New("could not read the network config")
	}
	if len(networkName) < 1 {
		c.NetworkName = networkInfo[0].Name
		c.Subnet = networkInfo[0].Subnet
		c.NakedSubnet = strings.Split(networkInfo[0].Subnet, "/")[1]
		c.Gateway = networkInfo[0].Gateway
		c.NetworkComment = networkInfo[0].Comment
	} else {
		for _, v := range networkInfo {
			if networkName == v.Name {
				c.NetworkName = v.Name
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

	if len(deployDnsServer) > 1 {
		c.DnsServer = deployDnsServer
	} else {
		c.DnsServer = c.Gateway
	}

	c.LiveStatus = "testing"
	c.OsType = "custom"
	c.OsComment = "Custom OS"

	c.ParentHost = GetHostName()
	c.VncPort = generateRandomVncPort()
	c.VncPassword = generateRandomPassword(8, true, true)
	if err != nil {
		return errors.New("could not generate vnc port: " + err.Error())
	}

	c.SshKeys, err = getSystemSshKeys()
	if err != nil {
		return errors.New("could not get ssh keys: " + err.Error())
	}

	// Generate template ciUserDataTemplate
	tmpl, err := template.New("ciUserDataTemplate").Parse(ciUserDataTemplate)
	if err != nil {
		return errors.New("could not generate ciUserDataTemplate: " + err.Error())
	}

	var ciUserData strings.Builder
	if err := tmpl.Execute(&ciUserData, c); err != nil {
		return errors.New("could not generate ciUserDataTemplate: " + err.Error())
	}

	// Generate template ciNetworkConfigTemplate
	tmpl, err = template.New("ciNetworkConfigTemplate").Parse(ciNetworkConfigTemplate)
	if err != nil {
		return errors.New("could not generate ciNetworkConfigTemplate: " + err.Error())
	}

	var ciNetworkConfig strings.Builder
	if err := tmpl.Execute(&ciNetworkConfig, c); err != nil {
		return errors.New("could not generate ciNetworkConfigTemplate: " + err.Error())
	}

	// Generate template ciMetaDataTemplate
	tmpl, err = template.New("ciMetaDataTemplate").Parse(ciMetaDataTemplate)
	if err != nil {
		return errors.New("could not generate ciMetaDataTemplate: " + err.Error())
	}

	var ciMetaData strings.Builder
	if err := tmpl.Execute(&ciMetaData, c); err != nil {
		return errors.New("could not generate ciMetaDataTemplate: " + err.Error())
	}

	// Move this into a separate function with the proper error handling
	zfsCreateOutput, zfsCreateErr := exec.Command("zfs", "create", dsParent+"/"+c.VmName).CombinedOutput()
	if zfsCreateErr != nil {
		return errors.New(strings.TrimSpace(string(zfsCreateOutput)) + zfsCreateErr.Error())
	}
	osDiskLocation := "/" + dsParent + "/" + c.VmName + "/disk0.img"
	_ = exec.Command("touch", osDiskLocation).Run()
	_ = exec.Command("truncate", "-s", "+10G", osDiskLocation).Run()
	emojlog.PrintLogMessage("Created a new VM dataset: "+dsParent+"/"+c.VmName, emojlog.Debug)

	// Write config files
	emojlog.PrintLogMessage("Writing CloudInit config files", emojlog.Debug)
	newVmFolder := "/" + dsParent + "/" + c.VmName
	vmConfigFileLocation := newVmFolder + "/vm_config.json"
	vmConfig := VmConfigStruct{}
	networkConfig := VmNetworkStruct{}
	diskConfig := VmDiskStruct{}
	diskConfigList := []VmDiskStruct{}
	vmConfig.CPUSockets = "1"
	vmConfig.CPUCores = c.Cpus
	vmConfig.Memory = c.Ram
	vmConfig.Loader = "uefi"
	vmConfig.LiveStatus = c.LiveStatus
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

	// Add system disk
	diskConfig.DiskType = "nvme"
	diskConfig.DiskLocation = "internal"
	diskConfig.DiskImage = "disk0.img"
	diskConfig.Comment = "OS Disk"
	diskConfigList = append(diskConfigList, diskConfig)
	// Add CloudInit ISO
	diskConfig.DiskType = "ahci-cd"
	diskConfig.DiskLocation = "internal"
	diskConfig.DiskImage = "seed.iso"
	diskConfig.Comment = "CloudInit ISO"
	diskConfigList = append(diskConfigList, diskConfig)
	// Add the installation ISO
	diskConfig.DiskType = "ahci-cd"
	diskConfig.DiskLocation = "external"
	diskConfig.DiskImage = isoPath
	diskConfig.Comment = "Installation ISO"
	diskConfigList = append(diskConfigList, diskConfig)
	// Translate the temp diskConfig variable into the struct
	vmConfig.Disks = append(vmConfig.Disks, diskConfigList...)

	vmConfig.IncludeHostwideSSHKeys = true
	vmConfig.VmSshKeys = c.SshKeys
	vmConfig.VncPort = c.VncPort
	vmConfig.VncPassword = c.VncPassword
	vmConfig.Description = "-"

	err = vmConfigFileWriter(vmConfig, vmConfigFileLocation)
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

	err = ReloadDnsServer()
	if err != nil {
		return err
	}

	// Start the VM when all of the above is complete
	if startWhenReady {
		time.Sleep(time.Second * 1)
		err := VmStart(c.VmName, false, true, true)
		if err != nil {
			return err
		}
	}

	return nil
}

const ciUserDataTemplate = `#cloud-config

users:
  - default
  - name: root
    lock_passwd: false
    ssh_pwauth: true
    disable_root: false
    ssh_authorized_keys:
	  {{- range .SshKeys}}
      - {{ .KeyValue }}
	  {{- end }}

  - name: gwitsuper
    sudo: ALL=(ALL) NOPASSWD:ALL
    groups: wheel
    ssh_pwauth: true
    lock_passwd: false
    ssh_authorized_keys:
	  {{- range .SshKeys}}
      - {{ .KeyValue }}
	  {{- end }}

chpasswd:
  list: |
    root:{{ .RootPassword }}
    gwitsuper:{{ .GwitsuperPassword }}
  expire: False

package_update: false
package_upgrade: false
`

const ciMetaDataTemplate = `instance-id: iid-{{ .InstanceId }}
local-hostname: {{ .VmName }}
`

const ciNetworkConfigTemplate = `version: 2
ethernets:
  interface0:
    match:
      macaddress: "{{ .MacAddress }}"

    {{- if not (or (eq .OsType "freebsd13ufs") (eq .OsType "freebsd13zfs")) }} 
    set-name: eth0
	{{- end }}
    addresses:
    - {{ .IpAddress }}/{{ .NakedSubnet }}
 
    gateway4: {{ .Gateway }}
 
    nameservers:
      search: [ {{ .ParentHost }}.internal.lan, ]
      addresses: [{{ .DnsServer }}, ]
`

func checkVmNameInput(vmName string) (vmNameError error) {
	vmNameMinLength := 5
	vmNameMaxLength := 22
	vmNameCantStartWith := "1234567890-_"
	vmNameValidChars := "qwertyuiopasdfghjklzxcvbnm-QWERTYUIOPASDFGHJKLZXCVBNM_1234567890"

	// Check if vmName uses valid characters
	for _, v := range vmName {
		valid := false
		for _, vv := range vmNameValidChars {
			if v == vv {
				valid = true
				break
			}
		}
		if !valid {
			vmNameError = errors.New("name cannot contain '" + string(v) + "' character")
			return
		}
	}
	// EOF Check if vmName uses valid characters

	// Check if vmName starts with a valid character
	for i, v := range vmName {
		if i > 1 {
			break
		}
		for _, vv := range vmNameCantStartWith {
			if v == vv {
				vmNameError = errors.New("name cannot start with a number, an underscore or a hyphen")
				return
			}
		}
	}
	// EOF Check if vmName starts with a valid character

	// Check vmName length
	if len(vmName) < vmNameMinLength {
		vmNameError = errors.New("name cannot contain less than " + strconv.Itoa(vmNameMinLength) + " characters")
		return
	} else if len(vmName) > vmNameMaxLength {
		vmNameError = errors.New("name cannot contain more than " + strconv.Itoa(vmNameMaxLength) + " characters")
		return
	}
	// EOF Check vmName length

	return
}

func generateNewIp(networkName string) (string, error) {
	var existingIps []string

	// Add existing VM IPs
	for _, v := range getAllVms() {
		networkNameFound := false
		tempConfig := vmConfig(v)
		for i, v := range tempConfig.Networks {
			if v.NetworkBridge == networkName {
				networkNameFound = true
				existingIps = append(existingIps, tempConfig.Networks[i].IPAddress)
			}
		}
		if !networkNameFound {
			existingIps = append(existingIps, tempConfig.Networks[0].IPAddress)
		}
	}
	// EOF Add existing VM IPs

	// Add existing Jail IPs
	jailList, err := GetAllJailsList()
	if err != nil {
		return "", err
	}
	for _, v := range jailList {
		jailsConfig, err := GetJailConfig(v, true)
		if err != nil {
			return "", nil
		}
		existingIps = append(existingIps, jailsConfig.IPAddress)
	}
	// EOF Add existing Jail IPs

	networks, err := networkInfo()
	if err != nil {
		return "", errors.New("could not read the config file: " + err.Error())
	}

	var subnet string
	var rangeStart string
	var rangeEnd string
	networkNameFoundGlobal := false
	for i, v := range networks {
		if v.Name == networkName {
			networkNameFoundGlobal = true
			subnet = networks[i].Subnet
			rangeStart = networks[i].RangeStart
			rangeEnd = networks[i].RangeEnd
		}
	}
	if !networkNameFoundGlobal {
		subnet = networks[0].Subnet
		rangeStart = networks[0].RangeStart
		rangeEnd = networks[0].RangeEnd
	}

	var randomIp string
	// var err error
	randomIp, err = generateUniqueRandomIp(subnet)
	if err != nil {
		return "", errors.New("could not generate a random IP address: " + err.Error())
	}

	iteration := 0
	for {
		if slices.Contains(existingIps, randomIp) || !ipIsWithinRange(randomIp, subnet, rangeStart, rangeEnd) {
			randomIp, err = generateUniqueRandomIp(subnet)
			if err != nil {
				return "", errors.New("could not generate a random IP address: " + err.Error())
			}
			iteration = iteration + 1
			if iteration > 400 {
				return "", errors.New("ran out of IP available addresses within this range")
			}
		} else {
			break
		}
	}

	return randomIp, nil
}

func generateUniqueRandomIp(subnet string) (string, error) {
	// Set the seed for the random number generator
	rand.Seed(time.Now().UnixNano())

	// Parse the subnet IP and mask
	ip, ipNet, err := net.ParseCIDR(subnet)
	if err != nil {
		return "", errors.New(err.Error())
	}

	// Calculate the size of the address space within the subnet
	size, _ := ipNet.Mask.Size()
	numHosts := (1 << (32 - size)) - 2

	// Generate a random host address within the subnet
	host := rand.Intn(numHosts) + 1
	addr := ip.Mask(ipNet.Mask)
	addr[0] |= byte(host >> 24)
	addr[1] |= byte(host >> 16)
	addr[2] |= byte(host >> 8)
	addr[3] |= byte(host)

	stringAddress := fmt.Sprintf("%v", addr)
	return stringAddress, nil
}

func ipIsWithinRange(ipAddress string, subnet string, rangeStart string, rangeEnd string) bool {
	// Parse the subnet IP and mask
	_, ipNet, err := net.ParseCIDR(subnet)
	if err != nil {
		panic(err)
	}

	// Define the range of allowed host addresses
	start := net.ParseIP(rangeStart).To4()
	end := net.ParseIP(rangeEnd).To4()

	// Parse the IP address to check
	ip := net.ParseIP(ipAddress).To4()

	// Check if the IP address is within the allowed range
	if ipNet.Contains(ip) && bytesInRange(ip, start, end) {
		return true
	} else {
		return false
	}
}

func bytesInRange(ip, start, end []byte) bool {
	for i := 0; i < len(ip); i++ {
		if start[i] > end[i] {
			log.Fatal("Make sure range start is lower than range end!")
		} else if ip[i] < start[i] || ip[i] > end[i] {
			return false
		}
	}
	return true
}

type NetworkInfoSt struct {
	Name            string `json:"network_name"`
	Gateway         string `json:"network_gateway"`
	Subnet          string `json:"network_subnet"`
	RangeStart      string `json:"network_range_start"`
	RangeEnd        string `json:"network_range_end"`
	BridgeInterface string `json:"bridge_interface"`
	ApplyBridgeAddr bool   `json:"apply_bridge_address"`
	Comment         string `json:"comment"`
}

func networkInfo() ([]NetworkInfoSt, error) {
	// JSON config file location
	execPath, err := os.Executable()
	if err != nil {
		return []NetworkInfoSt{}, err
	}
	networkConfigFile := path.Dir(execPath) + "/config_files/network_config.json"

	// Read the JSON file
	data, err := os.ReadFile(networkConfigFile)
	if err != nil {
		return []NetworkInfoSt{}, err
	}

	// Unmarshal the JSON data into a slice of Network structs
	var networks []NetworkInfoSt
	err = json.Unmarshal(data, &networks)
	if err != nil {
		return []NetworkInfoSt{}, err
	}

	return networks, nil
}

func generateRandomMacAddress() (string, error) {
	var existingMacs []string
	for _, v := range getAllVms() {
		tempConfig := vmConfig(v)
		existingMacs = append(existingMacs, tempConfig.Networks[0].NetworkMac)
	}

	macStr := ""
	for {
		if slices.Contains(existingMacs, macStr) || len(macStr) < 1 {
			// Generate a random MAC address
			mac := make([]byte, 3)
			_, err := rand.Read(mac)
			if err != nil {
				return "", err
			}

			// Format the MAC address as a string with the desired prefix
			macStr = fmt.Sprintf("58:9c:fc:%02x:%02x:%02x", mac[0], mac[1], mac[2])
		} else {
			break
		}
	}

	return macStr, nil
}

func generateVmName(vmName string) (string, error) {
	reAllowed := regexp.MustCompile(`[^a-zA-Z0-9\-]`)
	iter := 1
	vms := getAllVms()
	if reAllowed.MatchString(vmName) {
		return "", errors.New("name can only include A-Z, dash (-), and/or numbers")
	} else if string(vmName[len(vmName)-1]) == "-" {
		return "", errors.New("name cannot end with a dash (-)")
	} else if vmName == "test-vm" {
		vmName = "test-vm-" + strconv.Itoa(iter)
		for {
			if slices.Contains(vms, vmName) {
				iter = iter + 1
				vmName = "test-vm-" + strconv.Itoa(iter)
			} else {
				break
			}
		}
	} else if slices.Contains(vms, vmName) {
		return "", errors.New("vm already exists")
	}
	return vmName, nil
}

// Generate a random password given the length and character types
func generateRandomPassword(length int, caps, nums bool) string {
	// Define the character set for the password
	charset := "abcdefghijklmnopqrstuvwxyz"
	capS := "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	numS := "0123456789"
	if caps {
		charset = charset + capS
	}
	if nums {
		charset = charset + numS
	}

	rand.Seed(time.Now().UnixNano())
	result := ""
	iter := 0
	for {
		pwByte := charset[rand.Intn(len(charset))]
		result = result + string(pwByte)
		iter = iter + 1
		if iter > length {
			break
		}
	}

	return result
}

func generateRandomVncPort() string {
	var existingPorts []string
	startPort := 5900
	endPort := 6300
	for _, v := range getAllVms() {
		tempConfig := vmConfig(v)
		existingPorts = append(existingPorts, tempConfig.VncPort)
	}
	for {
		if slices.Contains(existingPorts, strconv.Itoa(startPort)) {
			startPort = startPort + 1
			continue
		} else if startPort > endPort {
			startPort = 5900
		} else {
			break
		}
	}

	return strconv.Itoa(startPort)
}

type HostConfigKey struct {
	KeyValue string `json:"key_value"`
	Comment  string `json:"comment"`
}

type HostConfig struct {
	ImageServer    string          `json:"public_vm_image_server"`
	BackupServers  []string        `json:"backup_servers"`
	ActiveDatasets []string        `json:"active_datasets"`
	DnsServers     []string        `json:"dns_servers,omitempty"`   // this is a new field, that might not be implemented on all of the nodes yet
	HostDNSACLs    []string        `json:"host_dns_acls,omitempty"` // this field is deprecated, remains here for backwards compatibility, and will be removed at some point
	HostSSHKeys    []HostConfigKey `json:"host_ssh_keys"`
}

func GetHostConfig() (HostConfig, error) {
	hostConfig := HostConfig{}
	execPath, err := os.Executable()
	if err != nil {
		return HostConfig{}, err
	}
	hostConfigFile := path.Dir(execPath) + "/config_files/host_config.json"
	data, err := os.ReadFile(hostConfigFile)
	if err != nil {
		return HostConfig{}, err
	}
	err = json.Unmarshal(data, &hostConfig)
	if err != nil {
		return HostConfig{}, err
	}
	return hostConfig, nil
}

func getSystemSshKeys() ([]VmSshKey, error) {
	sshKeys := []VmSshKey{}
	hostConfig := HostConfig{}
	// JSON config file location
	execPath, err := os.Executable()
	if err != nil {
		return sshKeys, err
	}
	hostConfigFile := path.Dir(execPath) + "/config_files/host_config.json"

	// Read the JSON file
	data, err := os.ReadFile(hostConfigFile)
	if err != nil {
		return sshKeys, err
	}

	// Unmarshal the JSON data into a slice of Network structs
	err = json.Unmarshal(data, &hostConfig)
	if err != nil {
		return sshKeys, err
	}

	for _, v := range hostConfig.HostSSHKeys {
		tempKey := VmSshKey{}
		tempKey.KeyValue = v.KeyValue
		tempKey.Comment = v.Comment
		tempKey.KeyOwner = "System"
		sshKeys = append(sshKeys, tempKey)
	}

	return sshKeys, nil
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

	snapName := "@deployment_" + newVmName + "_" + generateRandomPassword(8, false, true)
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

	emojlog.PrintLogMessage("New CloudInit ISO has been created", emojlog.Info)
	return nil
}

func vmConfigFileWriter(vmConfig VmConfigStruct, configLocation string) error {
	vmFileJsonOutput, err := json.MarshalIndent(vmConfig, "", "   ")
	if err != nil {
		return err
	}

	// Open the file in write-only mode, truncating (overwriting) it if it already exists
	file, err := os.OpenFile(configLocation, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	// Write data to the file
	_, err = file.Write(vmFileJsonOutput)
	if err != nil {
		return err
	}

	return nil
}
