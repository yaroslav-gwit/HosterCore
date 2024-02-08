package HosterJail

import (
	FreeBSDOsInfo "HosterCore/internal/pkg/freebsd/info"
	FreeBSDsysctls "HosterCore/internal/pkg/freebsd/sysctls"
	HosterHost "HosterCore/internal/pkg/hoster/host"
	HosterHostUtils "HosterCore/internal/pkg/hoster/host/utils"
	HosterJailUtils "HosterCore/internal/pkg/hoster/jail/utils"
	HosterNetwork "HosterCore/internal/pkg/hoster/network"
	HosterVmUtils "HosterCore/internal/pkg/hoster/vm/utils"
	HosterZfs "HosterCore/internal/pkg/hoster/zfs"
	"fmt"
	"html/template"
	"os"
	"strconv"
	"strings"
)

type DeployInput struct {
	JailName  string
	DsParent  string
	Release   string
	CpuLimit  int
	RamLimit  string
	IpAddress string
	Network   string
	DnsServer string
}

func Deploy(input DeployInput) error {
	// If the logger was already set, ignore this
	if !log.ConfigSet {
		log.SetFileLocation(HosterJailUtils.JAIL_AUDIT_LOG_LOCATION)
	}
	var err error
	prod := true

	// Generate a test-jail-{number} name automatically, if none was given
	if len(strings.TrimSpace(input.JailName)) < 1 {
		input.JailName, err = generateJailTestName()
		prod = false
		if err != nil {
			return err
		}
	}
	log.Info("Deploying a new Jail: " + input.JailName)

	err = HosterVmUtils.ValidateResName(input.JailName)
	if err != nil {
		return err
	}

	if len(input.DsParent) < 1 {
		hostCfg, err := HosterHost.GetHostConfig()
		if err != nil {
			return err
		}

		input.DsParent = hostCfg.ActiveZfsDatasets[0]
	}

	if len(input.Release) < 1 {
		input.Release, err = FreeBSDOsInfo.GetMajorReleaseVersion()
		if err != nil {
			return err
		}
	}

	jailConfig, err := generateJailDeployConfig(input.CpuLimit, input.RamLimit, input.IpAddress, input.Network, input.DnsServer, prod)
	if err != nil {
		return err
	}

	err = HosterJailUtils.ZfsTemplateClone(input.JailName, input.DsParent, input.Release)
	if err != nil {
		return err
	}

	// Create jail_config.json
	templateJail, err := template.New("templateJailConfigJson").Parse(HosterJailUtils.TemplateJailConfigJson)
	if err != nil {
		return err
	}
	fileTemplateJail, err := os.Create(fmt.Sprintf("/%s/%s/jail_config.json", input.DsParent, input.JailName))
	if err != nil {
		return err
	}
	err = templateJail.Execute(fileTemplateJail, jailConfig)
	if err != nil {
		return err
	}
	// EOF Create jail_config.json

	// Create jail_custom_parameters.conf
	err = os.WriteFile(fmt.Sprintf("/%s/%s/jail_custom_parameters.conf", input.DsParent, input.JailName), []byte(HosterJailUtils.TemplateJailCustomParameters), 0640)
	if err != nil {
		return err
	}
	// EOF Create jail_custom_parameters.conf

	log.Info("Jail has been deployed: " + input.JailName)
	return nil
}

func generateJailTestName() (r string, e error) {
	allDatasets, err := HosterZfs.ListMountPoints()
	if err != nil {
		e = err
		return
	}
	activeDatasets, err := HosterHost.GetHostConfig()
	if err != nil {
		e = err
		return
	}

	var existingFolders []string
	for _, v := range activeDatasets.ActiveZfsDatasets {
		for _, vv := range allDatasets {
			if vv.Mountpoint == "-" {
				continue
			}
			if vv.DsName != v {
				continue
			}

			entries, err := os.ReadDir(vv.Mountpoint + "/")
			if err != nil {
				e = fmt.Errorf("could not list files in the directory: %s", err.Error())
				return
			}
			for _, vv := range entries {
				existingFolders = append(existingFolders, vv.Name())
			}

		}
	}

	jailId := 1
	r = "test-jail-"
	tempJailName := r + strconv.Itoa(jailId)

jailNameLoop:
	for {
		foundJail := false

		for _, v := range existingFolders {
			if v == tempJailName {
				foundJail = true
				break
			} else {
				continue
			}
		}

		if foundJail {
			jailId += 1
			tempJailName = r + strconv.Itoa(jailId)
			continue jailNameLoop
		} else {
			r = tempJailName
			// break jailNameLoop
			return
		}
	}
}

func generateJailDeployConfig(cpuLimit int, ramLimit string, ipAddress string, network string, dnsServer string, prod bool) (r HosterJailUtils.JailConfig, e error) {
	r.CPULimitPercent = cpuLimit
	r.RAMLimit = ramLimit

	networks, err := HosterNetwork.GetNetworkConfig()
	if err != nil {
		e = err
		return
	}

	networkFound := false
	networkIndex := 0
	if len(network) < 1 {
		r.Network = networks[0].NetworkName
		r.DnsServer = networks[0].Gateway
	} else {
		for i, v := range networks {
			if v.NetworkName == network {
				networkFound = true
				networkIndex = i
			}
		}
		if networkFound {
			r.Network = network
			r.DnsServer = networks[networkIndex].Gateway
		} else {
			e = fmt.Errorf("network %s could not be found", network)
			return
		}
	}

	if len(ipAddress) < 1 {
		r.IPAddress, err = HosterHostUtils.GenerateNewRandomIp(r.Network)
		if err != nil {
			e = err
			return
		}
	} else {
		r.IPAddress = ipAddress
	}

	if len(dnsServer) > 0 {
		r.DnsServer = dnsServer
	}

	r.Timezone = "Europe/London"
	r.Parent, _ = FreeBSDsysctls.SysctlKernHostname()
	r.Production = prod
	r.Description = "-"

	return
}
