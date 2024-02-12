// Copyright 2023 Hoster Authors. All rights reserved.
// Use of this source code is governed by an Apache License 2.0
// license that can be found in the LICENSE file.

package HosterVmUtils

import (
	FileExists "HosterCore/internal/pkg/file_exists"
	"strings"
)

// Function takes in a folder path and checks if the jail configuration file exists in it. If it does, it will return true.
//
// For example: func(/zroot/vm-encrypted/test-vm-1) or func(/zroot/vm-encrypted/test-vm-1/)
//
// NOTE: Trailing "/" is automatically removed.
func VmConfigExists(folderPath string) (r bool) {
	// Remove the trailing "/" in the path if it exists
	folderPath = strings.TrimSuffix(folderPath, "/")

	// Return true if the config file was found
	if FileExists.CheckUsingOsStat(folderPath + "/" + VM_CONFIG_NAME) {
		r = true
		return
	}

	// Return false by default, if the config was not found
	return
}
