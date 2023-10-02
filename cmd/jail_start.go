package cmd

import (
	"HosterCore/emojlog"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"text/template"

	"github.com/spf13/cobra"
)

var (
	jailStartCmd = &cobra.Command{
		Use:   "start [jailName]",
		Short: "Start a specific Jail",
		Long:  `Start a specific Jail using it's name`,
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			err := checkInitFile()
			if err != nil {
				emojlog.PrintLogMessage(err.Error(), emojlog.Error)
				os.Exit(1)
			}

			err = jailStart(args[0], true)
			if err != nil {
				emojlog.PrintLogMessage(err.Error(), emojlog.Error)
				os.Exit(1)
			}
		},
	}
)

func jailStart(jailName string, logActions bool) error {
	jailConfig, err := getJailConfig(jailName, false)
	if err != nil {
		return err
	}

	t, err := template.New("templateJailRunningConfigPartial").Parse(templateJailRunningConfigPartial)
	if err != nil {
		return err
	}

	var jailConfigBuffer bytes.Buffer
	err = t.Execute(&jailConfigBuffer, jailConfig)
	if err != nil {
		return err
	}

	var jailConfigString = jailConfigBuffer.String()

	var additionalConfig []byte
	if FileExists(jailConfig.JailFolder + jailConfig.ConfigFileAppend) {
		additionalConfig, err = os.ReadFile(jailConfig.JailFolder + jailConfig.ConfigFileAppend)
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

	if logActions {
		emojlog.PrintLogMessage("Starting the Jail: "+jailName, emojlog.Info)
	}

	err = createMissingConfigFiles(jailConfig)
	if err != nil {
		return err
	}

	_ = os.Remove(jailConfig.JailFolder + "jail_temp_runtime.conf")
	err = os.WriteFile(jailConfig.JailFolder+"jail_temp_runtime.conf", []byte(jailConfigString), 0644)
	if err != nil {
		return err
	}

	_ = os.Remove("/etc/jail.conf")
	err = os.WriteFile("/etc/jail.conf", []byte(jailConfigString), 0644)
	if err != nil {
		return err
	}

	emojlog.PrintLogMessage(fmt.Sprintf("Starting a jail %s. Please give it a moment...", jailName), emojlog.Debug)
	out, err := exec.Command("service", "jail", "onestart", jailName).CombinedOutput()
	if err != nil {
		errorValue := "FATAL: " + strings.TrimSpace(string(out)) + "; " + err.Error()
		return fmt.Errorf("%s", errorValue)
	}

	// jailCreateCommand := []string{"jail", "-c", "-f", jailConfig.JailFolder + "jail_temp_runtime.conf"}
	// if logActions {
	// 	emojlog.PrintLogMessage("Executing the Jail startup script: "+strings.Join(jailCreateCommand, " "), emojlog.Debug)
	// 	emojlog.PrintLogMessage("Please give it a moment...", emojlog.Debug)
	// }
	// jailCreateOutput, err := exec.Command("jail", "-c", "-f", jailConfig.JailFolder+"jail_temp_runtime.conf").CombinedOutput()
	// if err != nil {
	// 	errorValue := "FATAL: " + strings.TrimSpace(string(jailCreateOutput)) + "; " + err.Error()
	// 	return errors.New(errorValue)
	// }

	createJailUptimeStateFile(jailName)
	_ = os.Remove("/etc/jail.conf")

	if logActions {
		emojlog.PrintLogMessage("Created a Jail uptime state file", emojlog.Changed)
		emojlog.PrintLogMessage("The Jail is up now: "+jailName, emojlog.Changed)
	}

	return nil
}

const templateJailRunningConfigPartial = `# Running Jail config generated by Hoster
{{ .JailName }} {
    host.hostname = {{ .JailHostname }};
    ip4.addr = "vm-{{ .Network }}|{{ .IPAddress }}/{{ .Netmask }}";
    path = "{{ .JailRootPath }}";
    exec.start = "{{ .StartupScript }}";
    exec.stop = "{{ .ShutdownScript }}";
    exec.consolelog = "/var/log/jail_console_{{ .JailName }}.log";

    # Log Jail startup and shutdown
    exec.prestart = "logger HOSTER_JAILS starting the Jail: {{ .JailName }}";
    exec.poststart = "logger HOSTER_JAILS the Jail has been started: {{ .JailName }}";
    exec.prestop = "logger HOSTER_JAILS stopping the Jail: {{ .JailName }}";
    exec.poststop = "logger HOSTER_JAILS the Jail has been stopped: {{ .JailName }}";

    # Apply Jail resource limits
    exec.poststart += "rctl -a jail:{{ .JailName }}:vmemoryuse:deny={{ .RAMLimit }}";
    exec.poststart += "rctl -a jail:{{ .JailName }}:pcpu:deny={{ .CpuLimitReal }}";
    exec.poststop += "rctl -r jail:{{ .JailName }}";
    # exec.poststop += "umount {{ .JailRootPath }}/dev || logger 'HOSTER_JAILS could not unmount {{ .JailRootPath }}/dev'";

    # Apply timezone settings
    exec.poststart += "tzsetup -sC {{ .JailRootPath }} {{ .Timezone }}";

    exec.clean;
    stop.timeout = 10;

    # Additional config
`

func checkIfJailExists(jailName string) (jailExists bool) {
	datasets, err := getZfsDatasetInfo()
	if err != nil {
		return
	}

	for _, v := range datasets {
		if FileExists(v.MountPoint + "/" + jailName + "/jail_config.json") {
			return true
		}
	}

	return
}

func getJailConfig(jailName string, ignoreJailExistsCheck bool) (jailConfig JailConfigFileStruct, configError error) {
	if !ignoreJailExistsCheck {
		if !checkIfJailExists(jailName) {
			configError = errors.New("jail doesn't exist")
			return
		}
	}

	datasets, err := getZfsDatasetInfo()
	if err != nil {
		configError = err
		return
	}

	configFile := ""
	for _, v := range datasets {
		if FileExists(v.MountPoint + "/" + jailName + "/jail_config.json") {
			configFile = v.MountPoint + "/" + jailName + "/jail_config.json"
			jailConfig.JailFolder = v.MountPoint + "/" + jailName + "/"
			jailConfig.ZfsDatasetPath = v.Name + "/" + jailName
			jailConfig.JailRootPath = v.MountPoint + "/" + jailName + "/root_folder"
		}
	}

	configFileRead, err := os.ReadFile(configFile)
	if err != nil {
		configError = err
		return
	}

	unmarshalErr := json.Unmarshal(configFileRead, &jailConfig)
	if unmarshalErr != nil {
		configError = unmarshalErr
		return
	}

	jailConfig.JailName = jailName
	jailConfig.JailHostname = jailName + "." + GetHostName() + ".internal.lan"

	networks, err := networkInfo()
	if err != nil {
		configError = unmarshalErr
		return
	}

	networkFound := false
	for _, v := range networks {
		if v.Name == jailConfig.Network {
			networkFound = true
			reSplitAtSlash := regexp.MustCompile(`\/`)
			jailConfig.Netmask = reSplitAtSlash.Split(v.Subnet, -1)[1]
		}
	}
	if !networkFound {
		configError = errors.New("network " + jailConfig.Network + " was not found")
		return
	}

	commandOutput, err := exec.Command("sysctl", "-nq", "hw.ncpu").CombinedOutput()
	if err != nil {
		fmt.Println("Error", err.Error())
	}

	numberOfCpusInt, err := strconv.Atoi(strings.TrimSpace(string(commandOutput)))
	if err != nil {
		configError = err
		return
	}

	realCpuLimit := jailConfig.CPULimitPercent * numberOfCpusInt
	jailConfig.CpuLimitReal = realCpuLimit

	if len(jailConfig.Description) < 1 {
		jailConfig.Description = "0"
	}

	return
}

const templateJailRcConf = `# Hoster generated RC.CONF
clear_tmp_enable="YES"
syslogd_flags="-ss"
sendmail_enable="NONE"
sendmail_enable="NO"
sendmail_msp_queue_enable="NO"
`

const templateJailResolvConf = `# Hoster generated RESOLV.CONF
search {{ .Parent }}.internal.lan
nameserver {{ .DnsServer }}
`

func createMissingConfigFiles(jailConfig JailConfigFileStruct) error {
	if !FileExists(jailConfig.JailRootPath + "/etc/fstab") {
		_, _ = os.Create(jailConfig.JailRootPath + "/etc/fstab")
	}

	// rc.conf
	if !FileExists(jailConfig.JailRootPath + "/etc/rc.conf") {
		t, err := template.New("templateJailRcConf").Parse(templateJailRcConf)
		if err != nil {
			return err
		}

		file, err := os.Create(jailConfig.JailRootPath + "/etc/rc.conf")
		if err != nil {
			return err
		}

		err = t.Execute(file, jailConfig)
		if err != nil {
			file.Close()
			return err
		}

		file.Close()
	}

	// resolv.conf
	templateResolvConf, err := template.New("templateJailResolvConf").Parse(templateJailResolvConf)
	if err != nil {
		return err
	}

	fileResolvConf, err := os.Create(jailConfig.JailRootPath + "/etc/resolv.conf")
	if err != nil {
		return err
	}

	err = templateResolvConf.Execute(fileResolvConf, jailConfig)
	if err != nil {
		fileResolvConf.Close()
		return err
	}

	fileResolvConf.Close()
	return nil
}
