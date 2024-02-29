// Copyright 2024 Hoster Authors. All rights reserved.
// Use of this source code is governed by an Apache License 2.0
// license that can be found in the LICENSE file.

package HosterVmUtils

import (
	HosterJailUtils "HosterCore/internal/pkg/hoster/jail/utils"
	"fmt"
	"slices"
)

func GenerateTestVmName(vmName string) (r string, e error) {
	err := ValidateResName(vmName)
	if err != nil {
		e = err
		return
	}

	existing := []string{}
	vms, err := ListAllSimple()
	if err != nil {
		e = err
		return
	}
	jails, err := HosterJailUtils.ListAllSimple()
	if err != nil {
		e = err
		return
	}

	iter := 1
	r = fmt.Sprintf("test-vm-%d", iter)
	if vmName == "test-vm" || len(vmName) < 1 {
		for _, v := range vms {
			existing = append(existing, v.VmName)
		}
		for _, v := range jails {
			existing = append(existing, v.JailName)
		}

		for {
			if slices.Contains(existing, r) {
				iter += 1
				r = fmt.Sprintf("test-vm-%d", iter)
			} else {
				break
			}
		}
	}

	return
}
