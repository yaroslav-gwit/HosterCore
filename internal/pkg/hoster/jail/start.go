package HosterJail

import (
	FileExists "HosterCore/internal/pkg/file_exists"
	FreeBSDsysctls "HosterCore/internal/pkg/freebsd/sysctls"
	HosterJailUtils "HosterCore/internal/pkg/hoster/jail/utils"
	HosterNetwork "HosterCore/internal/pkg/hoster/network"
	HosterLogger "HosterCore/internal/pkg/logger"
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
	HosterLogger.SetFileLocation("/var/log/hoster_audit_jail.log")
	HosterLogger.Info("Starting the Jail: " + jailName)

	running, err := isJailRunning(jailName)
	if err != nil {
		return err
	}
	if running {
		//lint:ignore ST1005 ignore this!
		return errors.New("Jail is already running")
	}

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
		//lint:ignore ST1005 ignore this!
		return fmt.Errorf("Jail doesn't exist: %s", jailName)
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

	err = createMissingConfigFiles(jailConfig, jailDsFolder+"/"+JAIL_ROOT_FOLDER)
	if err != nil {
		return err
	}

	jailStartConf, err := setJailStartValues(jailName, jailDsFolder, jailConfig, ifaces)
	if err != nil {
		return err
	}

	jailTempRuntimeLocation, err := generatePartialTemplate(jailStartConf, jailConfig, jailDsFolder)
	if err != nil {
		return err
	}

	out, err := exec.Command("jail", "-f", jailTempRuntimeLocation, "-c").CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s; %s", strings.TrimSpace(string(out)), err.Error())
	}

	err = HosterJailUtils.CreateUptimeStateFile(jailName)
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
		t, err := template.New("rc.conf").Parse(HosterJailUtils.TemplateJailRcConf)
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
	templateResolvConf, err := template.New("resolv.conf").Parse(HosterJailUtils.TemplateJailResolvConf)
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

func isJailRunning(jailName string) (r bool, e error) {
	jailsOnline, err := GetRunningJails()
	if err != nil {
		e = err
		return
	}

	for _, v := range jailsOnline {
		if v.Name == jailName {
			r = true
			return
		}
	}

	return
}

func setJailStartValues(jailName string, jailDsFolder string, jailConfig JailConfig, ifaces HosterNetwork.EpairInterface) (r JailStart, e error) {
	r.JailConfig = jailConfig
	r.JailName = jailName

	hostname, _ := FreeBSDsysctls.SysctlKernHostname()
	r.JailHostname = jailName + "." + hostname + "." + "lan"

	r.JailRootPath = jailDsFolder + "/" + JAIL_ROOT_FOLDER
	cpus, err := FreeBSDsysctls.SysctlHwVmmMaxcpu()
	if err != nil {
		e = err
		return
	}
	r.CpuLimitReal = jailConfig.CPULimitPercent * cpus
	r.EpairInterface = ifaces

	networks, err := HosterNetwork.GetNetworkConfig()
	if err != nil {
		e = err
		return
	}

	for _, v := range networks {
		if jailConfig.Network == v.NetworkName {
			r.DefaultRouter = v.Gateway
			Netmask := strings.Split(v.Subnet, "/")[1]
			r.Netmask = Netmask
		}
	}

	return
}

func generatePartialTemplate(jailStartConf JailStart, jailConfig JailConfig, jailDsFolder string) (r string, e error) {
	t, err := template.New("jConfigPartial").Parse(HosterJailUtils.TemplateJailRunningConfigPartial)
	if err != nil {
		e = err
		return
	}

	var jailConfigBuffer bytes.Buffer
	err = t.Execute(&jailConfigBuffer, jailStartConf)
	if err != nil {
		e = err
		return
	}
	var jailConfigString = jailConfigBuffer.String()

	var additionalConfig []byte
	if FileExists.CheckUsingOsStat(jailDsFolder + "/" + jailConfig.ConfigFileAppend) {
		additionalConfig, err = os.ReadFile(jailDsFolder + "/" + jailConfig.ConfigFileAppend)
		if err != nil {
			e = err
			return
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

	// Generate and write the Jail runtime config file
	var jailTempRuntimeLocation = jailDsFolder + "/" + JAIL_TEMP_RUNTIME
	r = jailTempRuntimeLocation
	_ = os.Remove(jailTempRuntimeLocation)
	err = os.WriteFile(jailTempRuntimeLocation, []byte(jailConfigString), 0640)
	if err != nil {
		e = err
		return
	}
	// EOF Generate and write the Jail runtime config file

	return
}
