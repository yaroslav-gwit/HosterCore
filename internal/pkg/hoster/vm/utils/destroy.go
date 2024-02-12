// Copyright 2023 Hoster Authors. All rights reserved.
// Use of this source code is governed by an Apache License 2.0
// license that can be found in the LICENSE file.

package HosterVmUtils

import (
	"fmt"
	"os/exec"
	"slices"
	"strings"
)

func Destroy(vmName string) error {
	running, err := GetRunningVms()
	if err != nil {
		return err
	}
	if slices.Contains(running, vmName) {
		return fmt.Errorf("vm is running")
	}

	vms, err := ListAllSimple()
	if err != nil {
		return err
	}

	vmFound := false
	vmInfo := VmListSimple{}
	for _, v := range vms {
		if v.VmName == vmName {
			vmInfo = v
			vmFound = true
		}
	}
	if !vmFound {
		return fmt.Errorf("vm doesn't exist")
	}

	out, err := exec.Command("zfs", "destroy", "-r", vmInfo.DsName+"/"+vmName).CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s; %s", strings.TrimSpace(string(out)), err.Error())
	}

	return nil
}
