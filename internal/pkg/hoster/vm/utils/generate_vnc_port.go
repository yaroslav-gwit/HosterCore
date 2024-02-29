// Copyright 2024 Hoster Authors. All rights reserved.
// Use of this source code is governed by an Apache License 2.0
// license that can be found in the LICENSE file.

package HosterVmUtils

import (
	"fmt"
	"slices"
)

func GenerateVncPort() (r int, e error) {
	maxPort := 6300
	r = 5900

	vms, err := ListJsonApi()
	if err != nil {
		e = err
		return
	}

	existing := []int{}
	for _, v := range vms {
		existing = append(existing, v.VncPort)
	}

	for {
		if r > maxPort {
			e = fmt.Errorf("ran out of available VNC ports")
			return
		}

		if slices.Contains(existing, r) {
			r += 1
		} else {
			break
		}
	}

	return
}
