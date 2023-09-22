package cmd

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"text/template"

	"github.com/spf13/cobra"
)

var (
	jailStartCmd = &cobra.Command{
		Use:   "start",
		Short: "Start a specific Jail",
		Long:  `Start a specific Jail using it's name`,
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			err := checkInitFile()
			if err != nil {
				log.Fatal(err.Error())
			}
			// cmd.Help()

			err = jailStart(args[0])
			if err != nil {
				log.Fatal(err.Error())
			}
		},
	}
)

func jailStart(jailName string) error {
	jailConfig, err := getJailConfig(jailName)
	if err != nil {
		return err
	}

	t, err := template.New("jailRunningConfigPartialTemplate").Parse(jailRunningConfigPartialTemplate)
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
		jailConfigString = jailConfigString + string(additionalConfig)
	}

	jailConfigString = jailConfigString + "\n}"
	fmt.Println(jailConfigString)

	return nil
}

const jailRunningConfigPartialTemplate = `# Running Jail config generated by Hoster
{{ .JailName }} {
    host.hostname = {{ .JailHostname }};
    ip4.addr = "vm-{{ .Network }}|{{ .IPAddress }}/{{ .Netmask }}";
    path = "{{ .JailRootPath }}";
    exec.clean;
    exec.start = "{{ .StartupScript }}";
    exec.stop = "{{ .ShutdownScript }}";
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

func getJailConfig(jailName string) (jailConfig JailConfigFileStruct, configError error) {
	if !checkIfJailExists(jailName) {
		configError = errors.New("jail doesn't exist")
		return
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
	jailConfig.Netmask = "24"

	return
}