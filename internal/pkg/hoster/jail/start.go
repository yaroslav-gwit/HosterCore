package HosterJail

import (
	FileExists "HosterCore/internal/pkg/file_exists"
	FreeBSDsysctls "HosterCore/internal/pkg/freebsd/sysctls"
	HosterNetwork "HosterCore/internal/pkg/hoster/network"
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"os"
	"os/exec"
	"strings"
)

type JailStart struct {
	JailName      string
	JailHostname  string
	JailRootPath  string
	CpuLimitReal  int
	DefaultRouter string
	Netmask       string
	JailConfig
	HosterNetwork.EpairInterface
}

func Start(jailName string) error {
	// Check if Jail is already online
	jailsOnline, err := GetRunningJails()
	if err != nil {
		return err
	}
	for _, v := range jailsOnline {
		if v.Name == jailName {
			return errors.New("this Jail is already running: " + jailName)
		}
	}
	// EOF Check if Jail is already online

	// Check if Jail exists and get it's dataset configuration
	jails, err := ListAllSimple()
	if err != nil {
		return err
	}
	jailDsInfo := JailListSimple{}
	jailFound := false
	for _, v := range jails {
		if v.JailName == jailName {
			jailFound = true
			jailDsInfo = v
		}
	}
	if !jailFound {
		return fmt.Errorf("this Jail was not found: %s", jailName)
	}
	jailDsFolder := jailDsInfo.MountPoint.Mountpoint + "/" + jailName
	// EOF Check if Jail exists and get it's dataset configuration

	jailConfig, err := GetJailConfig(jailDsFolder)
	if err != nil {
		return err
	}

	ifaces, err := HosterNetwork.CreateEpairInterface(jailName, jailConfig.Network)
	if err != nil {
		return err
	}

	// Set JailStart values
	jailStartConf := JailStart{}
	jailStartConf.JailConfig = jailConfig
	jailStartConf.JailName = jailName
	hostname, _ := FreeBSDsysctls.SysctlKernHostname()
	jailStartConf.JailHostname = jailName + "." + hostname + "." + "lan"
	jailStartConf.JailRootPath = jailDsFolder + "/" + JAIL_ROOT_FOLDER
	cpus, err := FreeBSDsysctls.SysctlHwVmmMaxcpu()
	if err != nil {
		return err
	}
	jailStartConf.CpuLimitReal = jailConfig.CPULimitPercent * cpus
	jailStartConf.EpairInterface = ifaces

	networks, err := HosterNetwork.GetNetworkConfig()
	if err != nil {
		return err
	}
	for _, v := range networks {
		if jailConfig.Network == v.NetworkName {
			jailStartConf.DefaultRouter = v.Gateway
			Netmask := strings.Split(v.Subnet, "/")[1]
			jailStartConf.Netmask = Netmask
		}
	}
	// EOF Set JailStart values

	t, err := template.New("templateJailRunningConfigPartial").Parse(templateJailRunningConfigPartial)
	if err != nil {
		return err
	}

	var jailConfigBuffer bytes.Buffer
	err = t.Execute(&jailConfigBuffer, jailStartConf)
	if err != nil {
		return err
	}
	var jailConfigString = jailConfigBuffer.String()

	var additionalConfig []byte
	if FileExists.CheckUsingOsStat(jailDsFolder + "/" + jailConfig.ConfigFileAppend) {
		additionalConfig, err = os.ReadFile(jailDsFolder + "/" + jailConfig.ConfigFileAppend)
		if err != nil {
			return err
		}
	}

	if len(additionalConfig) > 0 {
		additionalConfigSplit := strings.Split(string(additionalConfig), "\n")
		for _, v := range additionalConfigSplit {
			if len(v) > 0 {
				v = strings.TrimSpace(v)
				jailConfigString = jailConfigString + "    " + v + "\n"
			}
		}
		jailConfigString = jailConfigString + "}"
	} else {
		jailConfigString = jailConfigString + "\n}"
	}

	err = createMissingConfigFiles(jailConfig, jailDsFolder+"/"+"root_folder")
	if err != nil {
		return err
	}

	// Generate and write the Jail runtime config file
	_ = os.Remove(jailDsFolder + "/" + "jail_temp_runtime.conf")
	err = os.WriteFile(jailDsFolder+"/"+"jail_temp_runtime.conf", []byte(jailConfigString), 0640)
	if err != nil {
		return err
	}
	// EOF Generate and write the Jail runtime config file

	// Execute Jail start shell command
	out, err := exec.Command("jail", "-f", jailDsFolder+"/"+"jail_temp_runtime.conf", "-c").CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s; %s", strings.TrimSpace(string(out)), err.Error())
	}
	// EOF Execute Jail start shell command

	err = CreateUptimeStateFile(jailName)
	if err != nil {
		return err
	}

	return nil
}

func createMissingConfigFiles(jailConfig JailConfig, jailRootPath string) error {
	if !FileExists.CheckUsingOsStat(jailRootPath + "/etc/fstab") {
		_, _ = os.Create(jailRootPath + "/etc/fstab")
	}

	// rc.conf
	if !FileExists.CheckUsingOsStat(jailRootPath + "/etc/rc.conf") {
		t, err := template.New("templateJailRcConf").Parse(templateJailRcConf)
		if err != nil {
			return err
		}

		file, err := os.Create(jailRootPath + "/etc/rc.conf")
		if err != nil {
			return err
		}
		defer file.Close()

		err = t.Execute(file, jailConfig)
		if err != nil {
			file.Close()
			return err
		}
	}
	// EOF rc.conf

	// resolv.conf
	templateResolvConf, err := template.New("templateJailResolvConf").Parse(templateJailResolvConf)
	if err != nil {
		return err
	}

	fileResolvConf, err := os.Create(jailRootPath + "/etc/resolv.conf")
	if err != nil {
		return err
	}

	err = templateResolvConf.Execute(fileResolvConf, jailConfig)
	if err != nil {
		fileResolvConf.Close()
		return err
	}
	fileResolvConf.Close()
	// EOF resolv.conf

	return nil
}
