package cmd

import (
	"bufio"
	"errors"
	"hoster/emojlog"
	"html/template"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/spf13/cobra"
)

var (
	vmDestroyCmd = &cobra.Command{
		Use:   "destroy",
		Short: "Destroy the VM",
		Long:  `Destroy the VM and it's parent snapshot (uses zfs destroy)`,
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			err := vmDestroy(args[0])
			if err != nil {
				log.Fatal("vmDestroy() error: " + err.Error())
			}
			err = generateNewDnsConfig()
			if err != nil {
				log.Fatal("generateNewDnsConfig() error: " + err.Error())
			}
			err = reloadDnsService()
			if err != nil {
				log.Fatal("reloadDnsService() error: " + err.Error())
			}
		},
	}
)

func vmDestroy(vmName string) error {
	vmDataset, err := getVmDataset(vmName)
	emojlog.PrintLogMessage("Destroying VM: "+vmName, emojlog.Info)
	emojlog.PrintLogMessage("Removing this VM dataset: "+vmDataset, emojlog.Changed)
	if err != nil {
		return errors.New("could not find the VM dataset: " + err.Error())
	}

	findZfsParentCmd := exec.Command("zfs", "list", "-Ho", "name,origin", vmDataset)
	reSplit := regexp.MustCompile(`\s+`)
	stdout, stderr := findZfsParentCmd.Output()
	if stderr != nil {
		return errors.New("could not execute zfs list: " + err.Error())
	}

	vmDatasetParent := reSplit.Split(string(stdout), -1)[1]
	emojlog.PrintLogMessage("Removing parent VM dataset: "+vmDatasetParent, emojlog.Changed)
	err = exec.Command("zfs", "destroy", "-r", vmDataset).Run()
	if stderr != nil {
		return errors.New("could not execute zfs destroy: " + err.Error())
	}
	err = exec.Command("zfs", "destroy", vmDatasetParent).Run()
	if stderr != nil {
		return errors.New("could not execute zfs destroy: " + err.Error())
	}

	return nil
}

const unboundConfigTemplate = `
# This file is automatically generated by HosterCore.
# Modifications will be overwritten!
server:
	username: unbound
	directory: /var/unbound
	chroot: /var/unbound
	
	pidfile: /var/run/local_unbound.pid
	auto-trust-anchor-file: /var/unbound/root.key
	
	interface: 0.0.0.0
	
	access-control: 127.0.0.0/8 allow
	access-control: 10.0.0.0/8 allow
	access-control: 172.16.0.0/12 allow
	access-control: 192.168.0.0/16 allow
	
	{{range .}}
	local-zone: "{{.VmName}}" redirect
	local-data: "{{.VmName}} A {{.IpAddress}}"
	{{end}}

include: /var/unbound/forward.conf
include: /var/unbound/lan-zones.conf
include: /var/unbound/control.conf
include: /var/unbound/conf.d/*.conf
`

type dnsInfoStruct struct {
	VmName    string
	IpAddress string
}

func generateNewDnsConfig() error {
	var vmDnsInfo []dnsInfoStruct

	for _, v := range getAllVms() {
		tempConfig := vmConfig(v)
		tempDnsInfo := dnsInfoStruct{}
		tempDnsInfo.IpAddress = tempConfig.Networks[0].IPAddress
		tempDnsInfo.VmName = v
		vmDnsInfo = append(vmDnsInfo, tempDnsInfo)
	}

	tmpl, err := template.New("config").Parse(unboundConfigTemplate)
	if err != nil {
		return errors.New("could not reload the unbound service: " + err.Error())
	}

	var output strings.Builder
	if err := tmpl.Execute(&output, vmDnsInfo); err != nil {
		return errors.New("could not reload the unbound service: " + err.Error())
	}
	// fmt.Println(output.String())

	// Open a new file for writing
	unboundConfigFile, err := os.Create("/var/unbound/unbound.conf")
	if err != nil {
		panic(err)
	}
	defer unboundConfigFile.Close()

	// Create a new writer
	writer := bufio.NewWriter(unboundConfigFile)

	// Write a string to the file
	str := output.String()
	_, err = writer.WriteString(str)
	if err != nil {
		return errors.New(err.Error())
	}

	// Flush the writer to ensure all data has been written to the file
	err = writer.Flush()
	if err != nil {
		return errors.New(err.Error())
	}

	emojlog.PrintLogMessage("Unbound service config file has been written successfully", emojlog.Changed)
	return nil
}

func reloadDnsService() error {
	err := exec.Command("service", "local_unbound", "reload").Run()
	if err != nil {
		return errors.New("could not reload the unbound service: " + err.Error())
	}

	emojlog.PrintLogMessage("DNS Service reloaded", emojlog.Info)
	return nil
}
