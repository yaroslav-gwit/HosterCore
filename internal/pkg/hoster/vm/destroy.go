// Copyright 2024 Hoster Authors. All rights reserved.
// Use of this source code is governed by an Apache License 2.0
// license that can be found in the LICENSE file.

package HosterVm

import (
	HosterVmUtils "HosterCore/internal/pkg/hoster/vm/utils"
	"fmt"
	"os/exec"
	"regexp"
	"slices"
	"strings"
)

// Destroy the VM using it's name. Returns an error if something went wrong.
func Destroy(vmName string) error {
	// If the logger was already set, ignore this
	if !log.ConfigSet {
		log.SetFileLocation(HosterVmUtils.VM_AUDIT_LOG_LOCATION)
	}

	running, err := HosterVmUtils.GetRunningVms()
	if err != nil {
		return err
	}
	if slices.Contains(running, vmName) {
		return fmt.Errorf("vm is running")
	}

	vms, err := HosterVmUtils.ListAllSimple()
	if err != nil {
		return err
	}

	vmFound := false
	vmInfo := HosterVmUtils.VmListSimple{}
	for _, v := range vms {
		if v.VmName == vmName {
			vmInfo = v
			vmFound = true
		}
	}
	if !vmFound {
		return fmt.Errorf("vm doesn't exist")
	}

	vmDataset := vmInfo.DsName + "/" + vmName
	// Get the parent dataset
	out, err := exec.Command("zfs", "get", "-H", "origin", vmDataset).CombinedOutput()
	// Correct value:
	// NAME                           PROPERTY  VALUE                                                                           SOURCE
	//    [0]                             [1]           [2]                                                                     [3]
	// zroot/vm-encrypted/twelveFour	origin	zroot/vm-encrypted/template-debian12@deployment_twelveFour_qc5q7u6khy	         -
	// Empty value
	// zroot/vm-encrypted/wordpress-one	origin	         -	                                                                     -
	if err != nil {
		errorValue := "could not find a parent DS: " + strings.TrimSpace(string(out)) + "; " + err.Error()
		return fmt.Errorf("%s", errorValue)
	}
	reSpaceSplit := regexp.MustCompile(`\s+`)
	parentDataset := reSpaceSplit.Split(strings.TrimSpace(string(out)), -1)[2]
	// EOF Get the parent dataset

	out, err = exec.Command("zfs", "destroy", "-r", vmDataset).CombinedOutput()
	if err != nil {
		errValue := fmt.Sprintf("%s; %s", strings.TrimSpace(string(out)), err.Error())
		log.Error("vm could not be destroyed: " + vmName + "; error: " + errValue)
		return fmt.Errorf(errValue)
	}

	// Remove the parent dataset if it exists
	reMatch := regexp.MustCompile(`deployment_`)
	if len(parentDataset) > 1 && reMatch.MatchString(parentDataset) {
		out, err := exec.Command("zfs", "destroy", parentDataset).CombinedOutput()
		if err != nil {
			errorValue := "FATAL: " + strings.TrimSpace(string(out)) + "; " + err.Error()
			return fmt.Errorf("%s", errorValue)
		}
		log.Warn("VM parent dataset has been destroyed: " + parentDataset)
	}
	// EOF Remove the parent dataset if it exists

	log.Warn("vm has been destroyed: " + vmName + "; dataset: " + vmDataset)
	return nil
}
