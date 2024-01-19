package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

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

func GetJailConfig(jailName string, ignoreJailExistsCheck bool) (jailConfig JailConfigFileStruct, configError error) {
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
	network := NetworkInfoSt{}
	for _, v := range networks {
		if v.Name == jailConfig.Network {
			networkFound = true
			network = v
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
	jailConfig.DefaultRouter = network.Gateway

	if len(jailConfig.Description) < 1 {
		jailConfig.Description = "0"
	}

	return
}
