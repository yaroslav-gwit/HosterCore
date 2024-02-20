// Copyright 2023 Hoster Authors. All rights reserved.
// Use of this source code is governed by an Apache License 2.0
// license that can be found in the LICENSE file.

package HosterJailUtils

import (
	FileExists "HosterCore/internal/pkg/file_exists"
	"encoding/json"
	"errors"
	"os"
	"strings"
)

type JailConfig struct {
	CPULimitPercent  int      `json:"cpu_limit_percent"`
	RAMLimit         string   `json:"ram_limit"`
	StartupScript    string   `json:"startup_script"`
	ShutdownScript   string   `json:"shutdown_script"`
	ConfigFileAppend string   `json:"config_file_append"`
	IPAddress        string   `json:"ip_address"`
	Network          string   `json:"network"`
	DnsServer        string   `json:"dns_server"`
	Timezone         string   `json:"timezone"`
	Parent           string   `json:"parent"`
	Production       bool     `json:"production"`
	Tags             []string `json:"tags"`
	Description      string   `json:"description"`
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

	return
}
