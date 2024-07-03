// Copyright 2024 Hoster Authors. All rights reserved.
// Use of this source code is governed by an Apache License 2.0
// license that can be found in the LICENSE file.

package HosterVmUtils

import (
	"fmt"
	"slices"
)

func GenerateVncPort() (r int, e error) {
	minPort := 5900
	maxPort := 7999

	vms, err := ListJsonApi()
	if err != nil {
		e = err
		return
	}

	var usedPorts []int
	for _, v := range vms {
		usedPorts = append(usedPorts, v.VncPort)
	}

	port := minPort
	for {
		if !slices.Contains(usedPorts, port) {
			r = port
			break
		}
		port++
	}

	if r > maxPort {
		e = fmt.Errorf("ran out of available VNC ports (tried ports %d to %d)", minPort, maxPort)
		return
	}

	return
}
