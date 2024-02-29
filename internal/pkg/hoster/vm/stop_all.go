// Copyright 2024 Hoster Authors. All rights reserved.
// Use of this source code is governed by an Apache License 2.0
// license that can be found in the LICENSE file.

package HosterVm

import (
	HosterVmUtils "HosterCore/internal/pkg/hoster/vm/utils"
	"slices"
)

// This function stops all running VMs by sending a specific `kill` signal to the underlying `bhyve` processes.
//
// Returns an error if something went wrong.
func StopAll(forceKill bool, forceCleanup bool) error {
	// If the logger was already set, ignore this
	if !log.ConfigSet {
		log.SetFileLocation(HosterVmUtils.VM_AUDIT_LOG_LOCATION)
	}

	vmsRunning, _ := HosterVmUtils.GetRunningVms()
	vms, err := HosterVmUtils.ListAllSimple()
	if err != nil {
		return err
	}

	for _, v := range vms {
		// Check if the VM is running block
		if !slices.Contains(vmsRunning, v.VmName) {
			continue
		}
		// EOF Check if the VM is running block

		err = SendShutdownSignal(v.VmName, forceKill, forceCleanup)
		if err != nil {
			return err
		}
	}

	return nil
}
