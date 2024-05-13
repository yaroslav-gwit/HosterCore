// Copyright 2024 Hoster Authors. All rights reserved.
// Use of this source code is governed by an Apache License 2.0
// license that can be found in the LICENSE file.

//go:build freebsd
// +build freebsd

package HosterVm

import (
	ErrorMappings "HosterCore/internal/app/rest_api_v2/pkg/error_mappings"
	FileExists "HosterCore/internal/pkg/file_exists"
	HosterLocations "HosterCore/internal/pkg/hoster/locations"
	HosterVmUtils "HosterCore/internal/pkg/hoster/vm/utils"
	"fmt"
	"os"
	"os/exec"
	"slices"
	"syscall"

	"github.com/sirupsen/logrus"
)

func Start(vmName string, waitVnc bool, debugRun bool) error {
	// If the logger was already set, ignore this
	if !log.ConfigSet {
		log.SetLevel(logrus.DebugLevel)
		log.SetFileLocation(HosterVmUtils.VM_AUDIT_LOG_LOCATION)
	}
	// EOF If the logger was already set, ignore this

	// VM is live check
	live, err := HosterVmUtils.GetRunningVms()
	if err != nil {
		return err
	}
	if slices.Contains(live, vmName) {
		return fmt.Errorf("vm is already running")
	}
	// EOF VM is live check

	// Check if VM exists
	vmInfo := HosterVmUtils.VmApi{}
	vms, err := HosterVmUtils.ListJsonApi()
	if err != nil {
		return err
	}
	found := false
	for _, v := range vms {
		if v.Name == vmName {
			vmInfo = v
			found = true
		}
	}
	if !found {
		return fmt.Errorf(ErrorMappings.VmDoesntExist.String())
	}
	// EOF Check if VM exists

	// Check if VM is a backup
	if vmInfo.Backup {
		return fmt.Errorf("cannot start a backup VM on this system (change the parent host first or execute cireset)")
	}
	// EOF Check if VM is a backup

	log.Info("starting the vm: " + vmName)
	vmLocation := vmInfo.Simple.Mountpoint + "/" + vmName
	bhyveCmd, err := HosterVmUtils.GenerateBhyveStartCmd(vmName, vmLocation, false, waitVnc)
	if err != nil {
		return err
	}
	os.Setenv("VM_START", bhyveCmd)
	os.Setenv("VM_NAME", vmName)
	os.Setenv("LOG_FILE", vmLocation+"/"+HosterVmUtils.VM_LOG_NAME)
	log.Debug("bhyve cmd: " + bhyveCmd)

	binaryLoc := ""
	for _, v := range HosterLocations.GetBinaryFolders() {
		loc := v + "/vm_supervisor_service"
		if FileExists.CheckUsingOsStat(loc) {
			binaryLoc = loc
		}
	}
	if len(binaryLoc) < 1 {
		return fmt.Errorf("vm_supervisor_service has not been found on your system")
	}

	if !debugRun {
		cmd := exec.Command(binaryLoc, "for", vmName)
		cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
		err = cmd.Start()
		if err != nil {
			return err
		}
		log.Info("vm is now up: " + vmName)
	}

	return nil
}
