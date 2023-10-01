package cmd

import (
	"HosterCore/emojlog"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"text/template"

	"github.com/spf13/cobra"
)

var (
	jailDeployCmdOsRelease string
	jailDeployCmdDataset   string
	jailDeployCmdJailName  string
	// jailDeployCmdCpuLimit    int
	// jailDeployCmdRamLimit    string
	// jailDeployCmdIpAddress   string
	// jailDeployCmdNetwork     string
	// jailDeployCmdDnsServer   string
	// jailDeployCmdTimezone    string
	// jailDeployCmdProduction  string
	// jailDeployCmdDescription string

	jailDeployCmd = &cobra.Command{
		Use:   "deploy",
		Short: "Deploy a new Jail",
		Long:  `Deploy a new Jail.`,
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			err := checkInitFile()
			if err != nil {
				emojlog.PrintLogMessage(err.Error(), emojlog.Error)
				os.Exit(1)
			}

			err = deployNewJail(jailDeployCmdJailName, jailDeployCmdDataset, jailDeployCmdOsRelease)
			if err != nil {
				emojlog.PrintLogMessage(err.Error(), emojlog.Error)
				os.Exit(1)
			}
		},
	}
)

const templateJailConfigJson = `{
    "cpu_limit_percent": {{ .CPULimitPercent }},
    "ram_limit": "{{ .RAMLimit }}",

    "startup_script": "/bin/sh /etc/rc",
    "shutdown_script": "/bin/sh /etc/rc.shutdown",
    "config_file_append": "jail_custom_parameters.conf",

    "ip_address": "{{ .IPAddress }}",
    "network": "{{ .Network }}",
    "dns_server": "{{ .DnsServer }}",

    "timezone": "{{ .Timezone }}",
    "parent": "{{ .Parent }}",
    "production": {{ .Production }},
    "description": "{{ .Description }}"
}
`

const templateJailCustomParameters = `mount.devfs;
allow.raw_sockets = "1";
allow.sysvipc = "1";
`

func deployNewJail(jailName string, dsParent string, release string) error {
	var err error

	if len(jailName) < 1 {
		return fmt.Errorf("jail name parameter cannot be empty")
	}

	if len(dsParent) < 1 {
		datasets, err := getZfsDatasetInfo()
		if err != nil {
			return err
		}
		dsParent = datasets[0].Name
	}

	if len(release) < 1 {
		release, err = getFreeBsdRelease()
		if err != nil {
			return err
		}
	}

	jailConfig, err := generateJailDeployConfig()
	if err != nil {
		return err
	}

	err = executeJailZfsClone(jailName, dsParent, release)
	if err != nil {
		return err
	}
	emojlog.PrintLogMessage(fmt.Sprintf("Closed ZFS based Jail template %s", release), emojlog.Changed)

	// Create jail_config.json
	templateJail, err := template.New("templateJailConfigJson").Parse(templateJailConfigJson)
	if err != nil {
		return err
	}
	fileTemplateJail, err := os.Create(fmt.Sprintf("/%s/%s/jail_config.json", dsParent, jailName))
	if err != nil {
		return err
	}
	err = templateJail.Execute(fileTemplateJail, jailConfig)
	if err != nil {
		return err
	}
	emojlog.PrintLogMessage(fmt.Sprintf("Created /%s/%s/jail_config.json", dsParent, jailName), emojlog.Changed)
	// EOF Create jail_config.json

	// Create jail_custom_parameters.conf
	err = os.WriteFile(fmt.Sprintf("/%s/%s/jail_custom_parameters.conf", dsParent, jailName), []byte(templateJailCustomParameters), 0640)
	if err != nil {
		return err
	}
	emojlog.PrintLogMessage(fmt.Sprintf("Created /%s/%s/jail_custom_parameters.json", dsParent, jailName), emojlog.Changed)
	// EOF Create jail_custom_parameters.conf

	emojlog.PrintLogMessage(fmt.Sprintf("New Jail has been deployed: %s", jailName), emojlog.Info)

	return nil
}

func generateJailDeployConfig() (jailConfig JailConfigFileStruct, jailError error) {
	jailConfig.CpuLimitReal = 50
	jailConfig.RAMLimit = "1G"

	networks, err := networkInfo()
	if err != nil {
		jailError = err
		return
	}
	jailConfig.Network = networks[0].Name

	jailConfig.IPAddress, err = generateNewIp(networks[0].Name)
	if err != nil {
		jailError = err
		return
	}

	jailConfig.DnsServer = networks[0].Gateway
	jailConfig.Timezone = "Europe/London"
	jailConfig.Parent = GetHostName()
	jailConfig.Production = true
	jailConfig.Description = "-"

	return
}

func executeJailZfsClone(jailName string, dsParent string, release string) error {
	dsExists, err := doesDatasetExist(fmt.Sprintf("%s/jail-template-%s", dsParent, release))
	if err != nil {
		return err
	}
	if !dsExists {
		return fmt.Errorf("parent dataset does not exist: %s/jail-template-%s", dsParent, release)
	}

	jailSnapshotName := dsParent + "/jail-template-" + release + "@deployment_" + jailName + "_" + generateRandomPassword(9, false, true)
	out, err := exec.Command("zfs", "snapshot", jailSnapshotName).CombinedOutput()
	if err != nil {
		return fmt.Errorf("could not execute zfs snapshot: %s; %s", strings.TrimSpace(string(out)), err.Error())
	}

	out, err = exec.Command("zfs", "clone", jailSnapshotName, dsParent+"/"+jailName).CombinedOutput()
	if err != nil {
		return fmt.Errorf("could not execute zfs clone: %s; %s", strings.TrimSpace(string(out)), err.Error())
	}

	return nil
}
