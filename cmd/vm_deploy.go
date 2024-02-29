package cmd

import (
	"HosterCore/internal/pkg/emojlog"
	FreeBSDsysctls "HosterCore/internal/pkg/freebsd/sysctls"
	HosterHost "HosterCore/internal/pkg/hoster/host"
	HosterHostUtils "HosterCore/internal/pkg/hoster/host/utils"
	HosterNetwork "HosterCore/internal/pkg/hoster/network"
	HosterVm "HosterCore/internal/pkg/hoster/vm"
	HosterVmUtils "HosterCore/internal/pkg/hoster/vm/utils"
	"bufio"
	"errors"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"text/template"
	"time"

	"github.com/spf13/cobra"
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

			if len(zfsDataset) < 1 {
				hostCfg, err := HosterHost.GetHostConfig()
				if err != nil {
					emojlog.PrintLogMessage(err.Error(), emojlog.Error)
					os.Exit(1)
				}

				zfsDataset = hostCfg.ActiveZfsDatasets[0]
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
	// LiveStatus        string
}

func deployVmMain(vmName string, networkName string, osType string, dsParent string, cpus int, ram string, startWhenReady bool) error {
	vmNameError := HosterVmUtils.ValidateResName(vmName)
	if vmNameError != nil {
		return vmNameError
	}

	// Initialize values
	var err error
	c := ConfigOutputStruct{}

	// Set CPU cores and RAM
	c.Cpus = cpus
	c.Ram = ram

	// Generate and set the root and gwitsuper users password
	c.RootPassword = HosterHostUtils.GenerateRandomPassword(33, true, true)
	if err != nil {
		return errors.New("could not generate random password for root user: " + err.Error())
	}
	c.GwitsuperPassword = HosterHostUtils.GenerateRandomPassword(33, true, true)
	if err != nil {
		return errors.New("could not generate random password for gwitsuper user: " + err.Error())
	}

	// Generate and set CI instance ID
	c.InstanceId = HosterHostUtils.GenerateRandomPassword(5, false, true)
	if err != nil {
		return errors.New("could not generate random instance id: " + err.Error())
	}

	// Generate correct VM name
	c.VmName, err = HosterVmUtils.GenerateTestVmName(vmName)
	if err != nil {
		return errors.New("could not generate vm name: " + err.Error())
	}

	emojlog.PrintLogMessage("Deploying new VM: "+c.VmName, emojlog.Info)

	// Generate and set random MAC address
	c.MacAddress, err = HosterVmUtils.GenerateMacAddress()
	if err != nil {
		return errors.New("could not generate vm name: " + err.Error())
	}

	if len(deployIpAddress) > 1 {
		c.IpAddress = deployIpAddress
	} else {
		// Generate and set random IP address (which is free in the pool of addresses)
		c.IpAddress, err = HosterHostUtils.GenerateNewRandomIp(networkName)
		if err != nil {
			return errors.New("could not generate the IP: " + err.Error())
		}
	}

	networkInfo, err := HosterNetwork.GetNetworkConfig()
	if err != nil {
		return errors.New("could not read the network config")
	}
	if len(networkName) < 1 {
		c.NetworkName = networkInfo[0].NetworkName
		c.Subnet = networkInfo[0].Subnet
		c.NakedSubnet = strings.Split(networkInfo[0].Subnet, "/")[1]
		c.Gateway = networkInfo[0].Gateway
		c.NetworkComment = networkInfo[0].Comment
	} else {
		for _, v := range networkInfo {
			if networkName == v.NetworkName {
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

	if len(deployDnsServer) > 1 {
		c.DnsServer = deployDnsServer
	} else {
		c.DnsServer = c.Gateway
	}

	reMatchTest := regexp.MustCompile(`.*test`)
	if reMatchTest.MatchString(c.VmName) {
		c.Production = false
	} else {
		c.Production = true
	}

	emojlog.PrintLogMessage("OS type used: "+osType, emojlog.Debug)
	c.OsType = osType
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

	err = ReloadDnsServer()
	if err != nil {
		return err
	}

	// Start the VM when all of the above is complete
	if startWhenReady {
		time.Sleep(time.Second * 1)
		// err := VmStart(c.VmName, false, false, false)
		err := HosterVm.Start(vmName, false, false)
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

	vmNameError := HosterVmUtils.ValidateResName(vmName)
	if vmNameError != nil {
		return vmNameError
	}

	// Initialize values
	c := ConfigOutputStruct{}
	var err error

	// Set CPU cores and RAM
	c.Cpus = cpus
	c.Ram = ram

	// Generate and set the root and gwitsuper users password
	c.RootPassword = HosterHostUtils.GenerateRandomPassword(33, true, true)
	c.GwitsuperPassword = HosterHostUtils.GenerateRandomPassword(33, true, true)
	// Generate and set CI instance ID
	c.InstanceId = HosterHostUtils.GenerateRandomPassword(5, false, true)

	// Generate correct VM name
	c.VmName, err = HosterVmUtils.GenerateTestVmName(vmName)
	if err != nil {
		return errors.New("could not set a vm name: " + err.Error())
	}

	emojlog.PrintLogMessage("Deploying new VM: "+c.VmName, emojlog.Info)

	// Generate and set random MAC address
	c.MacAddress, err = HosterVmUtils.GenerateMacAddress()
	if err != nil {
		return errors.New("could not generate vm name: " + err.Error())
	}

	if len(deployIpAddress) > 1 {
		c.IpAddress = deployIpAddress
	} else {
		// Generate and set random IP address (which is free in the pool of addresses)
		c.IpAddress, err = HosterHostUtils.GenerateNewRandomIp(networkName)
		if err != nil {
			return errors.New("could not generate the IP: " + err.Error())
		}
	}

	netInfo, err := HosterNetwork.GetNetworkConfig()
	if err != nil {
		return errors.New("could not read the network config")
	}
	if len(networkName) < 1 {
		c.NetworkName = netInfo[0].NetworkName
		c.Subnet = netInfo[0].Subnet
		c.NakedSubnet = strings.Split(netInfo[0].Subnet, "/")[1]
		c.Gateway = netInfo[0].Gateway
		c.NetworkComment = netInfo[0].Comment
	} else {
		for _, v := range netInfo {
			if networkName == v.NetworkName {
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

	if len(deployDnsServer) > 1 {
		c.DnsServer = deployDnsServer
	} else {
		c.DnsServer = c.Gateway
	}

	c.Production = false
	c.OsType = "custom"
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

	vmConfig.IncludeHostSSHKeys = true
	vmConfig.VmSshKeys = c.SshKeys
	vmConfig.VncPort = c.VncPort
	vmConfig.VncPassword = c.VncPassword
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

	err = ReloadDnsServer()
	if err != nil {
		return err
	}

	// Start the VM when all of the above is complete
	if startWhenReady {
		time.Sleep(time.Second * 1)
		// err := VmStart(c.VmName, false, true, true)
		err := HosterVm.Start(vmName, true, false)
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

	emojlog.PrintLogMessage("New CloudInit ISO has been created", emojlog.Info)
	return nil
}
