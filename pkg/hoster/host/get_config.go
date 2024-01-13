package HosterHost

import (
	FileExists "HosterCore/pkg/file_exists"
	"encoding/json"
	"errors"
	"os"
)

type HostConfigKey struct {
	KeyValue string `json:"key_value"`
	Comment  string `json:"comment"`
}

type HostConfig struct {
	ImageServer       string          `json:"public_vm_image_server"`
	ActiveZfsDatasets []string        `json:"active_datasets"`
	DnsServers        []string        `json:"dns_servers,omitempty"`
	HostSSHKeys       []HostConfigKey `json:"host_ssh_keys"`
}

// An internal function, that loops through the list of possible
// config locations and picks up the first one available.
//
// Used only in the GetHostConfig().
func getHostConfigLocation() (r string, e error) {
	configFiles := []string{
		"/opt/hoster-core/config_files/host_config.json",
		"/opt/hoster/config_files/host_config.json",
		"/usr/local/etc/hoster/host_config.json",
		"/etc/hoster/host_config.json",
		"/root/hoster/host_config.json",
	}

	for _, v := range configFiles {
		if FileExists.CheckUsingOsStat(v) {
			r = v
			return
		}
	}

	e = errors.New("could not find the config file")
	return
}

// Parses the host_config.json, and returns the underlying struct or an error
func GetHostConfig() (r HostConfig, e error) {
	hostConfigFile, err := getHostConfigLocation()
	if err != nil {
		e = err
		return
	}

	data, err := os.ReadFile(hostConfigFile)
	if err != nil {
		e = err
		return
	}

	err = json.Unmarshal(data, &r)
	if err != nil {
		e = err
		return
	}

	return
}
