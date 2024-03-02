package cmd

import (
	"HosterCore/internal/pkg/emojlog"
	FreeBSDsysctls "HosterCore/internal/pkg/freebsd/sysctls"
	HosterHostUtils "HosterCore/internal/pkg/hoster/host/utils"
	HosterNetwork "HosterCore/internal/pkg/hoster/network"
	HosterVmUtils "HosterCore/internal/pkg/hoster/vm/utils"
	"bufio"
	"errors"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"text/template"

	"github.com/spf13/cobra"
)

var (
	ciResetCmdNewVmName   string
	ciResetCmdNetworkName string
	ciResetCmdIpAddress   string
	ciResetCmdDnsServer   string

	vmCiResetCmd = &cobra.Command{
		Use:   "cireset [vmName]",
		Short: "Reset VM's passwords, ssh keys, and network config (useful after VM migration)",
		Long:  `Reset VM's passwords, ssh keys, and network config (useful after VM migration)`,
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			checkInitFile()

			err := CiReset(args[0], ciResetCmdNewVmName)
			if err != nil {
				emojlog.PrintLogMessage(err.Error(), emojlog.Error)
				os.Exit(1)
			}
		},
	}
)

func CiReset(oldVmName string, newVmName string) error {
	// Initialize values
	var err error
	c := ConfigOutputStruct{}
	vmConf, err := HosterVmUtils.InfoJsonApi(oldVmName)
	if err != nil {
		return err
	}
	if vmConf.Running {
		return errors.New("vm has to be stopped")
	}

	// Collect the required information
	c.RootPassword = HosterHostUtils.GenerateRandomPassword(33, true, true)
	c.GwitsuperPassword = HosterHostUtils.GenerateRandomPassword(33, true, true)
	c.InstanceId = HosterHostUtils.GenerateRandomPassword(5, false, true)

	if len(newVmName) > 0 {
		c.VmName = newVmName
	} else {
		c.VmName = oldVmName
	}

	c.MacAddress, err = HosterVmUtils.GenerateMacAddress()
	if err != nil {
		return errors.New("could not generate vm name: " + err.Error())
	}

	if len(ciResetCmdIpAddress) > 1 {
		c.IpAddress = ciResetCmdIpAddress
	} else {
		// Generate and set random IP address (which is free in the pool of addresses)
		c.IpAddress, err = HosterHostUtils.GenerateNewRandomIp(ciResetCmdNetworkName)
		if err != nil {
			return errors.New("could not generate the IP: " + err.Error())
		}
	}

	networkInfo, err := HosterNetwork.GetNetworkConfig()
	if err != nil {
		return errors.New("could not read the network config")
	}
	if len(ciResetCmdNetworkName) < 1 {
		c.NetworkName = networkInfo[0].NetworkName
		c.Subnet = networkInfo[0].Subnet
		c.NakedSubnet = strings.Split(networkInfo[0].Subnet, "/")[1]
		c.Gateway = networkInfo[0].Gateway
		c.NetworkComment = networkInfo[0].Comment
	} else {
		for _, v := range networkInfo {
			if ciResetCmdNetworkName == v.NetworkName {
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

	if len(ciResetCmdDnsServer) > 1 {
		c.DnsServer = ciResetCmdDnsServer
	} else {
		c.DnsServer = c.Gateway
	}

	c.Cpus = vmConf.VmConfig.CPUCores
	c.Ram = vmConf.VmConfig.Memory
	c.Production = vmConf.Production
	c.OsType = vmConf.OsType
	c.OsComment = vmConf.OsComment
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

	oldDsName, err := getVmDataset(oldVmName)
	if err != nil {
		return errors.New(err.Error())
	}

	_ = os.Remove("/" + oldDsName + "/seed.iso")
	_ = os.RemoveAll("/" + oldDsName + "/cloud-init-files")

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

	// var vmName string
	var newDsName string
	if len(newVmName) > 0 {
		reVmNameReplace := regexp.MustCompile(`/` + oldVmName + `$`)
		newDsName = reVmNameReplace.ReplaceAllString(oldDsName, "/"+newVmName)
		if err := zfsDsRename(oldDsName, newDsName); err != nil {
			return errors.New(err.Error())
		}
		// vmName = newVmName
	} else {
		// vmName = oldVmName
		newDsName = oldDsName
	}

	// Write config files
	newVmFolder := "/" + newDsName

	vmConf.Networks[0].NetworkBridge = c.NetworkName
	vmConf.Networks[0].NetworkMac = c.MacAddress
	vmConf.Networks[0].IPAddress = c.IpAddress
	vmConf.Networks[0].Comment = c.NetworkComment
	vmConf.ParentHost = c.ParentHost
	vmConf.VncPort = c.VncPort
	vmConf.VncPassword = c.VncPassword
	vmConf.VmSshKeys = c.SshKeys

	vmConfigFileLocation := newVmFolder + "/" + HosterVmUtils.VM_CONFIG_NAME
	err = HosterVmUtils.ConfigFileWriter(vmConf.VmConfig, vmConfigFileLocation)
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

	return nil
}

func zfsDsRename(oldDsName, newDsName string) error {
	err := exec.Command("zfs", "rename", oldDsName, newDsName).Run()
	if err != nil {
		return errors.New("could not execute zfs snapshot: " + err.Error())
	}
	return nil
}
