package main

import (
	FileExists "HosterCore/internal/pkg/file_exists"
	HosterLocations "HosterCore/internal/pkg/hoster/locations"
	"encoding/json"
	"errors"
	"os"
)

type RestApiConfig struct {
	Bind     string `json:"bind"`
	Port     int    `json:"port"`
	Protocol string `json:"protocol"`
	HaMode   bool   `json:"ha_mode"`
	HaDebug  bool   `json:"ha_debug"`
	HTTPAuth []struct {
		User     string `json:"user"`
		Password string `json:"password"`
		HaUser   bool   `json:"ha_user"`
	} `json:"http_auth"`
}

const confFileName = "restapi_config.json"

// An internal function, that loops through the list of possible
// config locations and picks up the first one available.
//
// Used only in the GetHostConfig().
func getApiConfigLocation() (r string, e error) {
	for _, v := range HosterLocations.GetConfigFolders() {
		configLocation := v + "/" + confFileName
		if FileExists.CheckUsingOsStat(configLocation) {
			r = configLocation
			return
		}
	}

	e = errors.New("could not find the config file: " + confFileName)
	return
}

// Parses the host_config.json, and returns the underlying struct or an error
func GetApiConfig() (r RestApiConfig, e error) {
	apiConfigFile, err := getApiConfigLocation()
	if err != nil {
		e = err
		return
	}

	data, err := os.ReadFile(apiConfigFile)
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
