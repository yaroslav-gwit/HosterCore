// Copyright 2024 Hoster Authors. All rights reserved.
// Use of this source code is governed by an Apache License 2.0
// license that can be found in the LICENSE file.

package HosterVmUtils

import (
	"encoding/json"
	"os"
)

// Function that writes a new config to the disk.
// It takes in a new config struct to be written, and config file location, e.g. /tank/vm-encrypted/test-vm-1/vm_config.json
func ConfigFileWriter(conf VmConfig, confLocation string) error {
	jsonOutput, err := json.MarshalIndent(conf, "", "   ")
	if err != nil {
		return err
	}

	// Open the file in write-only mode, truncating (overwriting) it if it already exists
	file, err := os.OpenFile(confLocation, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	// Write data to the file
	_, err = file.Write(jsonOutput)
	if err != nil {
		return err
	}

	return nil
}
