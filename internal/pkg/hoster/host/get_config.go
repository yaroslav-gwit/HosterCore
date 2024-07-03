// Copyright 2024 Hoster Authors. All rights reserved.
// Use of this source code is governed by an Apache License 2.0
// license that can be found in the LICENSE file.

package HosterHost

import (
	FileExists "HosterCore/internal/pkg/file_exists"
	HosterLocations "HosterCore/internal/pkg/hoster/locations"
	"encoding/json"
	"errors"
	"os"
)

type HostConfigKey struct {
	KeyValue string `json:"key_value"`
	Comment  string `json:"comment"`
}

type DnsStaticRecord struct {
	Domain string `json:"domain"` // The domain name, e.g. "example.com" or simply "example"
	Type   string `json:"type"`   // The record type, e.g. "A", "AAAA", "CNAME", "TXT", "MX", "NS", "SRV", "SOA", "PTR"
	Data   string `json:"data"`   // The record data, e.g. "192.168.120.1" for A record, "mail.example.com" for CNAME, etc.
}

type HostConfig struct {
	ImageServer       string            `json:"public_vm_image_server"`
	DnsSearchDomain   string            `json:"dns_search_domain,omitempty"`
	Tags              []string          `json:"tags"`
	ActiveZfsDatasets []string          `json:"active_datasets"`
	DnsServers        []string          `json:"dns_servers,omitempty"`
	DnsStaticRecords  []DnsStaticRecord `json:"dns_static_records,omitempty"`
	HostSSHKeys       []HostConfigKey   `json:"host_ssh_keys"`
}

const confFileName = "host_config.json"

// A function, that loops through the list of possible config locations and picks up the first one available.
func getHostConfigLocation() (r string, e error) {
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

func SaveHostConfig(config HostConfig) error {
	confFile, err := getHostConfigLocation()
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
