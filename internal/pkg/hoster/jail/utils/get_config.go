// Copyright 2023 Hoster Authors. All rights reserved.
// Use of this source code is governed by an Apache License 2.0
// license that can be found in the LICENSE file.

package HosterJailUtils

import (
	FileExists "HosterCore/internal/pkg/file_exists"
	HosterHost "HosterCore/internal/pkg/hoster/host"
	"encoding/json"
	"errors"
	"os"
	"strings"
)

type JailConfig struct {
	Production       bool     `json:"production"`
	CPULimitPercent  int      `json:"cpu_limit_percent"`
	RAMLimit         string   `json:"ram_limit"`
	FailoverStrategy string   `json:"failover_strategy,omitempty"` // Can only be set to one of the two values: "cireset" or "change_parent"
	StartupScript    string   `json:"startup_script"`
	ShutdownScript   string   `json:"shutdown_script"`
	ConfigFileAppend string   `json:"config_file_append"`
	IPAddress        string   `json:"ip_address"`
	Network          string   `json:"network"`
	DnsSearchDomain  string   `json:"dns_search_domain,omitempty"`
	DnsServer        string   `json:"dns_server"`
	Timezone         string   `json:"timezone"`
	Parent           string   `json:"parent"`
	UUID             string   `json:"uuid,omitempty"`
	Description      string   `json:"description"`
	Tags             []string `json:"tags"`
}

const jailConfFilename = "jail_config.json"

// Reads and returns the jail_config.json as Go struct.
//
// Takes in a jail location folder, similar to this: "/hast_shared/test-vm-1" or "/hast_shared/test-vm-1/" (trailing "/" automatically removed).
func GetJailConfig(jailLocation string) (r JailConfig, e error) {
	jailLocation = strings.TrimSuffix(jailLocation, "/")

	jailConfLocation := jailLocation + "/" + jailConfFilename
	if !FileExists.CheckUsingOsStat(jailConfLocation) {
		e = errors.New("jail config file could not be found here: " + jailConfLocation)
		return
	}

	data, err := os.ReadFile(jailConfLocation)
	if err != nil {
		e = err
		return
	}

	err = json.Unmarshal(data, &r)
	if err != nil {
		e = err
		return
	}

	// Set the default failover strategy
	if len(r.FailoverStrategy) < 1 {
		r.FailoverStrategy = "change_parent"
	}
	if r.FailoverStrategy != "cireset" && r.FailoverStrategy != "change_parent" {
		r.FailoverStrategy = "change_parent"
	}

	// Set the default DNS search domain
	if len(r.DnsSearchDomain) < 1 {
		hostConfig, err := HosterHost.GetHostConfig()
		if err != nil {
			e = err
			return
		}
		r.DnsSearchDomain = hostConfig.DnsSearchDomain
	}

	return
}
