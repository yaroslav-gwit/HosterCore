package RestApiConfig

import (
	FileExists "HosterCore/internal/pkg/file_exists"
	HosterLocations "HosterCore/internal/pkg/hoster/locations"
	"encoding/json"
	"errors"
	"os"
)

type HaConfig struct {
	NodeType         string   `json:"node_type"`
	FailOverStrategy string   `json:"failover_strategy"`
	FailOverTime     int64    `json:"failover_time"`
	BackupNode       bool     `json:"backup_node"`
	Candidates       []HaNode `json:"candidates"`
	StartupTime      int64    `json:"startup_time"`
}

type HaNode struct {
	Hostname         string `json:"hostname"`
	Protocol         string `json:"protocol"`
	Address          string `json:"address"`
	Port             string `json:"port"`
	User             string `json:"user"`
	Password         string `json:"password"`
	FailOverStrategy string `json:"failover_strategy"`
	FailOverTime     int64  `json:"failover_time"`
	BackupNode       bool   `json:"backup_node"`
	StartupTime      int64  `json:"startup_time"`
	Registered       bool   `json:"registered"`
	TimesFailed      int    `json:"times_failed"`
}

const haConfFileName = "ha_config.json"

// An internal function, that loops through the list of possible
// config locations and picks up the first one available.
//
// Used only in the GetHostConfig().
func getHaConfigLocation() (r string, e error) {
	for _, v := range HosterLocations.GetConfigFolders() {
		configLocation := v + "/" + haConfFileName
		if FileExists.CheckUsingOsStat(configLocation) {
			r = configLocation
			return
		}
	}

	e = errors.New("could not find the config file: " + haConfFileName)
	return
}

// Parses the host_config.json, and returns the underlying struct or an error
func GetHaConfig() (r HaConfig, e error) {
	haConfigFile, err := getHaConfigLocation()
	if err != nil {
		e = err
		return
	}

	data, err := os.ReadFile(haConfigFile)
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
