// Copyright 2024 Hoster Authors. All rights reserved.
// Use of this source code is governed by an Apache License 2.0
// license that can be found in the LICENSE file.

package HosterVmUtils

import (
	"crypto/rand"
	"fmt"
	"slices"
)

func GenerateMacAddress() (r string, e error) {
	var existing []string

	vms, err := ListJsonApi()
	if err != nil {
		return "", err
	}
	for _, v := range vms {
		for _, vv := range v.Networks {
			existing = append(existing, vv.NetworkMac)
		}
	}

	for {
		if slices.Contains(existing, r) || len(r) < 1 {
			// Generate a random MAC address
			mac := make([]byte, 3)
			_, err := rand.Read(mac)
			if err != nil {
				return "", err
			}

			// Format the MAC address as a string with the desired prefix
			r = fmt.Sprintf("58:9c:fc:%02x:%02x:%02x", mac[0], mac[1], mac[2])
		} else {
			break
		}
	}

	return
}
