// Copyright 2023 Hoster Authors. All rights reserved.
// Use of this source code is governed by an Apache License 2.0
// license that can be found in the LICENSE file.

package HosterZfs

import (
	"fmt"
	"os/exec"
	"regexp"
	"strings"
)

type MountPoint struct {
	DsName     string
	Mountpoint string // Example: /tank/vm-encrypted
	Encrypted  bool
}

// Lists all ZFS datasets, and extracts some additional information.
func ListMountPoints() (r []MountPoint, e error) {
	out, err := exec.Command("zfs", "list", "-o", "name,mountpoint,encryption").CombinedOutput()
	if err != nil {
		e = fmt.Errorf("%s; %s", strings.TrimSpace(string(out)), err.Error())
		return
	}
	// Example output
	//   [0]                                            [1]                                            [2]
	// NAME                                          MOUNTPOINT                                     ENCRYPTION
	// hast_shared                                   /hast_shared                                   off
	// tank/vm-encrypted                             /tank/vm-encrypted                             aes-256-gcm
	// tank/vm-encrypted/test-vm-3                   /tank/vm-encrypted/test-vm-3                   aes-256-gcm
	// tank/vm-unencrypted                           /tank/vm-unencrypted                           off

	reSplitSpace := regexp.MustCompile(`\s+`)
	for i, v := range strings.Split(string(out), "\n") {
		// skip the header
		if i == 0 {
			continue
		}
		// skip empty lines
		if len(v) < 1 {
			continue
		}

		split := reSplitSpace.Split(v, -1)
		// Guardrails from panicking, if the amount of items is less than 3
		if len(split) < 3 {
			continue
		}

		// Set the encryption flag
		encrypted := false
		if split[2] != "off" {
			encrypted = true
		}

		r = append(r, MountPoint{DsName: split[0], Mountpoint: split[1], Encrypted: encrypted})
	}

	return
}
