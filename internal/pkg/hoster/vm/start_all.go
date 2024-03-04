// Copyright 2024 Hoster Authors. All rights reserved.
// Use of this source code is governed by an Apache License 2.0
// license that can be found in the LICENSE file.

package HosterVm

import (
	FileExists "HosterCore/internal/pkg/file_exists"
	HosterLocations "HosterCore/internal/pkg/hoster/locations"
	HosterVmUtils "HosterCore/internal/pkg/hoster/vm/utils"
	"fmt"
	"os"
	"os/exec"
	"slices"
	"syscall"
	"time"

	"github.com/sirupsen/logrus"
)

func StartAll(prodOnly bool, waitTime int) error {
	// If the logger was already set, ignore this
	if !log.ConfigSet {
		log.SetLevel(logrus.DebugLevel)
		log.SetFileLocation(HosterVmUtils.VM_AUDIT_LOG_LOCATION)
	}

	liveVms, err := HosterVmUtils.GetRunningVms()
	if err != nil {
		return err
	}
	vms, err := HosterVmUtils.ListJsonApi()
	if err != nil {
		return err
	}

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

	startId := 0
	for _, v := range vms {
		if slices.Contains(liveVms, v.Name) {
			continue
		}
		if prodOnly && !v.Production {
			continue
		}
		if v.Backup {
			continue
		}

		if startId > 0 {
			time.Sleep(time.Duration(waitTime * int(time.Second)))
		}
		startId += 1

		log.Info("starting the vm: " + v.Name)

		vmLocation := v.Simple.Mountpoint + "/" + v.Name
		bhyveCmd, err := HosterVmUtils.GenerateBhyveStartCmd(v.Name, vmLocation, false, false)
		if err != nil {
			return err
		}

		os.Setenv("VM_START", bhyveCmd)
		os.Setenv("VM_NAME", v.Name)
		os.Setenv("LOG_FILE", vmLocation+"/"+HosterVmUtils.VM_LOG_NAME)
		log.Debug("bhyve cmd: " + bhyveCmd)

		cmd := exec.Command(binaryLoc, "for", v.Name)
		cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
		err = cmd.Start()
		if err != nil {
			return err
		}

		log.Info("vm is now up: " + v.Name)
	}

	return nil
}
