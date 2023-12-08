// Copyright 2023s The Hoster Authors. All rights reserved.
// Use of this source code is governed by a Apache License 2.0
// license that can be found in the LICENSE fill

package host

import (
	"HosterCore/models/host"
	"encoding/json"
	"os"
	"path"
)

const ConfigPath = "/config_files/host_config.json"

func GetHostConfig() (host.Config, error) {
	hostConfig := host.Config{}

	execPath, err := os.Executable()
	if err != nil {
		return host.Config{}, err
	}

	hostConfigFile := path.Dir(execPath) + ConfigPath
	data, err := os.ReadFile(hostConfigFile)
	if err != nil {
		return host.Config{}, err
	}

	err = json.Unmarshal(data, &hostConfig)
	if err != nil {
		return host.Config{}, err
	}

	return hostConfig, nil
}
