// Copyright 2023 Hoster Authors. All rights reserved.
// Use of this source code is governed by an Apache License 2.0
// license that can be found in the LICENSE file.

package HosterNetwork

import (
	FileExists "HosterCore/internal/pkg/file_exists"
	HosterLocations "HosterCore/internal/pkg/hoster/locations"
	"encoding/json"
	"errors"
	"os"
)

type NetworkConfig struct {
	NetworkName     string `json:"network_name"`
	Gateway         string `json:"network_gateway"`
	Subnet          string `json:"network_subnet"`
	RangeStart      string `json:"network_range_start"`
	RangeEnd        string `json:"network_range_end"`
	BridgeInterface string `json:"bridge_interface"`
	ApplyBridgeAddr bool   `json:"apply_bridge_address"`
	Comment         string `json:"comment"`
}

const confFileName = "network_config.json"

// An internal function, that loops through the list of possible
// config locations and picks up the first one available.
//
// Used only in the GetHostConfig().
func getNetworkConfigLocation() (r string, e error) {
	for _, v := range HosterLocations.GetConfigFolders() {
		configLocation := v + "/" + confFileName
		if FileExists.CheckUsingOsStat(configLocation) {
			r = configLocation
			return
		}
	}

	e = errors.New("could not find the config file")
	return
}

// Parses the host_config.json, and returns the underlying struct or an error
func GetNetworkConfig() (r []NetworkConfig, e error) {
	networkConfigFile, err := getNetworkConfigLocation()
	if err != nil {
		e = err
		return
	}

	data, err := os.ReadFile(networkConfigFile)
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

// Saves the network config to the network_config.json file by taking in the NetworkConfig struct.
func SaveNetworkConfig(config []NetworkConfig) error {
	confFile, err := getNetworkConfigLocation()
	if err != nil {
		return err
	}

	jsonData, err := json.MarshalIndent(config, "", "   ")
	if err != nil {
		return err
	}

	file, err := os.OpenFile(confFile, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.Write(jsonData)
	if err != nil {
		return err
	}

	return nil
}
