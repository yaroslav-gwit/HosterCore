// Copyright 2024 Hoster Authors. All rights reserved.
// Use of this source code is governed by an Apache License 2.0
// license that can be found in the LICENSE file.

package HosterVm

import (
	FreeBSDsysctls "HosterCore/internal/pkg/freebsd/sysctls"
	HosterVmUtils "HosterCore/internal/pkg/hoster/vm/utils"
	"errors"
	"fmt"
)

// Replaces the parent on a VM specified, if the newParent is passed as an empty string GetHostName() will be automatically used.
func ChangeParent(vmName string, newParent string, ignoreLiveCheck bool) error {
	// If the logger was already set, ignore this
	if !log.ConfigSet {
		log.SetFileLocation(HosterVmUtils.VM_AUDIT_LOG_LOCATION)
	}

	if len(vmName) < 1 {
		return errors.New("you must provide a VM name")
	}

	vms, err := HosterVmUtils.ListJsonApi()
	if err != nil {
		return err
	}

	found := false
	vmInfo := HosterVmUtils.VmApi{}
	for _, v := range vms {
		if v.Name == vmName {
			found = true
			vmInfo = v
		}
	}
	if !found {
		return errors.New("vm does not exist on this system")
	}

	if !ignoreLiveCheck {
		if vmInfo.Running {
			return errors.New("vm must be offline to perform this operation")
		}
	}

	if len(newParent) < 1 {
		newParent, _ = FreeBSDsysctls.SysctlKernHostname()
	}

	vmFolder := vmInfo.Simple.Mountpoint + "/" + vmName
	vmConf := vmInfo.VmConfig
	if vmConf.ParentHost == newParent {
		log.Debug("No changes applied, because the old parent value is the same as a new parent value")
		return nil
	}
	vmConf.ParentHost = newParent

	err = HosterVmUtils.ConfigFileWriter(vmConf, vmFolder+"/"+HosterVmUtils.VM_CONFIG_NAME)
	if err != nil {
		return err
	}

	log.Info(fmt.Sprintf("Parent host has been changed for %s to %s", vmName, newParent))
	return nil
}
