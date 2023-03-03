package cmd

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"hoster/emojlog"
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
	osType                 string
	zfsDataset             string
	vmDeployCpus           int
	vmDeployRam            string
	vmDeployStartWhenReady bool

	vmDeployCmd = &cobra.Command{
		Use:   "deploy",
		Short: "Deploy the VM, using a pre-defined template",
		Long:  `Deploy the VM, using a pre-defined template`,
		Run: func(cmd *cobra.Command, args []string) {
			err := deployVmMain(vmName, osType, zfsDataset, vmDeployCpus, vmDeployRam, vmDeployStartWhenReady)
			if err != nil {
				log.Fatal(err)
			}
		},
	}
)

type SshKey struct {
	Key     string
	Owner   string
	Comment string
}

type ConfigOutputStruct struct {
	Cpus              string
	Ram               string
	SshKeys           []SshKey
	RootPassword      string
	GwitsuperPassword string
	InstanceId        string
	VmName            string
	MacAddress        string
	IpAddress         string
	Subnet            string
	NakedSubnet       string
	Gateway           string
	LiveStatus        string
	OsType            string
	OsComment         string
	ParentHost        string
	VncPort           string
	VncPassword       string
}

func deployVmMain(vmName string, osType string, dsParent string, cpus int, ram string, startWhenReady bool) error {
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

	// Generate and set random IP address (which is free in the pool of addresses)
	c.IpAddress, err = generateNewIp()
	if err != nil {
		return errors.New("could not generate the IP")
	}

	networkInfo, err := networkInfo()
	if err != nil {
		return errors.New("could not generate the IP")
	}
	c.Subnet = networkInfo[0].Subnet
	c.NakedSubnet = strings.Split(networkInfo[0].Subnet, "/")[1]
	c.Gateway = networkInfo[0].Gateway

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
	case "ubuntu2004":
		c.OsComment = "Ubuntu 20.04"
	case "almalinux8":
		c.OsComment = "AlmaLinux 8"
	case "rockylinux8":
		c.OsComment = "RockyLinux 8"
	case "freebsd13ufs":
		c.OsComment = "FreeBSD 13 UFW"
	case "freebsd13zfs":
		c.OsComment = "FreeBSD 13 ZFS"
	case "windows10":
		c.OsComment = "Windows 10"
	case "windows11":
		c.OsComment = "Windows 11"
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

	// Generate template vmConfigFileTemplate
	tmpl, err = template.New("vmConfigFileTemplate").Parse(vmConfigFileTemplate)
	if err != nil {
		return errors.New("could not generate vmConfigFileTemplate: " + err.Error())
	}

	var vmConfigFile strings.Builder
	if err := tmpl.Execute(&vmConfigFile, c); err != nil {
		return errors.New("could not generate vmConfigFileTemplate: " + err.Error())
	}
	// fmt.Println(vmConfigFile.String())

	zfsCloneResult, err := zfsDatasetClone(dsParent, osType, c.VmName)
	if err != nil || !zfsCloneResult {
		return errors.New(err.Error())
	}

	// Write config files
	emojlog.PrintLogMessage("Writing CloudInit config files", emojlog.Debug)
	newVmFolder := "/" + dsParent + "/" + c.VmName

	// Open a new file for writing
	vmConfigFileLocation, err := os.Create(newVmFolder + "/vm_config.json")
	if err != nil {
		return errors.New(err.Error())
	}
	defer vmConfigFileLocation.Close()
	// Create a new writer
	writer := bufio.NewWriter(vmConfigFileLocation)
	// Write a string to the file
	str := vmConfigFile.String()
	_, err = writer.WriteString(str)
	if err != nil {
		return errors.New(err.Error())
	}
	// Flush the writer to ensure all data has been written to the file
	err = writer.Flush()
	if err != nil {
		return errors.New(err.Error())
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
	writer = bufio.NewWriter(ciUserDataFileLocation)
	// Write a string to the file
	str = ciUserData.String()
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

	err = generateNewDnsConfig()
	if err != nil {
		return errors.New(err.Error())
	}
	err = reloadDnsService()
	if err != nil {
		return errors.New(err.Error())
	}

	// Start the VM when all of the above is complete
	if startWhenReady {
		time.Sleep(time.Second * 1)
		err := vmStart(c.VmName)
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
      - {{ .Key }}
	  {{- end }}

  - name: gwitsuper
    sudo: ALL=(ALL) NOPASSWD:ALL
    groups: wheel
    ssh_pwauth: true
    lock_passwd: false
    ssh_authorized_keys:
	  {{- range .SshKeys}}
      - {{ .Key }}
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
     
     set-name: eth0
     addresses:
     - {{ .IpAddress }}/{{ .NakedSubnet }}
     
     gateway4: {{ .Gateway }}
     
     nameservers:
       search: [gateway-it.internal, ]
       addresses: [{{ .Gateway }}, ]
`

const vmConfigFileTemplate = `
{
    "cpu_sockets": "1",
    "cpu_cores": "{{ .Cpus }}",
    "memory": "{{ .Ram }}",
    "loader": "uefi",
    "live_status": "{{ .LiveStatus }}",
    "os_type": "{{ .OsType }}",
    "os_comment": "{{ .OsComment }}",
    "owner": "System",
    "parent_host": "{{ .ParentHost }}",

    "networks": [
        {
            "network_adaptor_type": "virtio-net",
            "network_bridge": "internal",
            "network_mac": "{{ .MacAddress }}",
            "ip_address": "{{ .IpAddress }}",
            "comment": "Internal Network"
        }
    ],

    "disks": [
        {
            "disk_type": "virtio-blk",
            "disk_location": "internal",
            "disk_image": "disk0.img",
            "comment": "OS Drive"
        },
        {
            "disk_type": "ahci-cd",
            "disk_location": "internal",
            "disk_image": "seed.iso",
            "comment": "Cloud Init ISO"
        }
    ],

    "include_hostwide_ssh_keys": true,
    "vm_ssh_keys": [
        {}
    ],

    "vnc_port": "{{ .VncPort }}",
    "vnc_password": "{{ .VncPassword }}",

    "description": "-"
}
`

func generateNewIp() (string, error) {
	var existingIps []string
	for _, v := range getAllVms() {
		tempConfig := vmConfig(v)
		existingIps = append(existingIps, tempConfig.Networks[0].IPAddress)
	}

	networks, err := networkInfo()
	if err != nil {
		return "", errors.New(err.Error())
	}

	subnet := networks[0].Subnet
	rangeStart := networks[0].RangeStart
	rangeEnd := networks[0].RangeEnd

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
	HostDNSACLs    []string        `json:"host_dns_acls"`
	HostSSHKeys    []HostConfigKey `json:"host_ssh_keys"`
}

func getSystemSshKeys() ([]SshKey, error) {
	sshKeys := []SshKey{}
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
		tempKey := SshKey{}
		tempKey.Key = v.KeyValue
		tempKey.Comment = v.Comment
		tempKey.Owner = "System"
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
	err = exec.Command("zfs", "snapshot", vmTemplate+snapName).Run()
	if err != nil {
		return false, errors.New("could not execute zfs snapshot: " + err.Error())
	}

	err = exec.Command("zfs", "clone", vmTemplate+snapName, dsParent+"/"+newVmName).Run()
	if err != nil {
		return false, errors.New("could not execute zfs clone: " + err.Error())
	}

	return true, nil
}

func createCiIso(vmFolder string) error {
	ciFolder := vmFolder + "/cloud-init-files/"
	err := exec.Command("genisoimage", "-output", vmFolder+"/seed.iso", "-volid", "cidata", "-joliet", "-rock", ciFolder+"user-data", ciFolder+"meta-data", ciFolder+"network-config").Run()
	if err != nil {
		return errors.New("there was a problem generating an ISO: " + err.Error())
	}

	emojlog.PrintLogMessage("New CloudInit ISO has been created", emojlog.Info)
	return nil
}
