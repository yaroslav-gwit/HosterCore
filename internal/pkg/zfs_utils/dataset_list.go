// Copyright 2023 Hoster Authors. All rights reserved.
// Use of this source code is governed by an Apache License 2.0
// license that can be found in the LICENSE file.

package zfsutils

import (
	"errors"
	"os/exec"
	"regexp"
	"strings"
)

// Simply returns a list of all available ZFS datasets, using a default ZFS list command.
//
// Example return ["hast_shared/test-vm-1", "tank/vm-encrypted/prometheus"]
func DefaultDatasetList() ([]string, error) {
	out, err := exec.Command("zfs", "list", "-p").CombinedOutput()
	if err != nil {
		errString := strings.TrimSpace(string(out)) + "; " + err.Error()
		return []string{}, errors.New(errString)
	}
	// Example output
	// NAME                                                  USED          AVAIL        REFER  MOUNTPOINT
	// hast_shared                                    12279562240   503787601920       122880  /hast_shared
	// hast_shared/jail-template-13.2-RELEASE          1455566848   503787601920   1455566848  /hast_shared/jail-template-13.2-RELEASE
	// hast_shared/template-debian12                   5371150336   503787601920   5371150336  /hast_shared/template-debian12
	// hast_shared/test-vm-1                           2769940480   503787601920   6087208960  /hast_shared/test-vm-1
	// hast_shared/test-vm-2                           2667876352   503787601920   5840850944  /hast_shared/test-vm-2
	// tank                                          577733734400  1353390936064       106496  /tank
	// tank/brick_1                                   11247562752  1353390936064  11247562752  /tank/brick_1

	reSplitSpace := regexp.MustCompile(`\s+`)
	result := []string{}
	for i, v := range strings.Split(string(out), "\n") {
		if i == 0 {
			continue
		}

		v = reSplitSpace.Split(v, -1)[0]
		v = strings.TrimSpace(v)
		result = append(result, v)
	}

	return result, nil
}
