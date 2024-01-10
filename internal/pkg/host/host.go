// Copyright 2023s The Hoster Authors. All rights reserved.
// Use of this source code is governed by a Apache License 2.0
// license that can be found in the LICENSE fill

package host

import (
	hostconfig "HosterCore/internal/models/host_config"
	"encoding/json"
	"os"
)

const ConfigPath = "/opt/hoster-core/config_files/host_config.json"

func GetHostConfig() (hostconfig.Config, error) {
	hostConfig := hostconfig.Config{}

	data, err := os.ReadFile(ConfigPath)
	if err != nil {
		return hostconfig.Config{}, err
	}

	err = json.Unmarshal(data, &hostConfig)
	if err != nil {
		return hostconfig.Config{}, err
	}

	return hostConfig, nil
}
