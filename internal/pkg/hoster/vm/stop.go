// Copyright 2024 Hoster Authors. All rights reserved.
// Use of this source code is governed by an Apache License 2.0
// license that can be found in the LICENSE file.

package HosterVm

import (
	FreeBSDKill "HosterCore/internal/pkg/freebsd/kill"
	FreeBSDPgrep "HosterCore/internal/pkg/freebsd/pgrep"
	HosterNetwork "HosterCore/internal/pkg/hoster/network"
	HosterVmUtils "HosterCore/internal/pkg/hoster/vm/utils"
	"errors"
	"fmt"
	"regexp"
	"slices"
)

// This function stops any particular VM in an async fashion by sending a specific `kill` signal to the underlying `bhyve` process.
//
// Returns an error if something went wrong.
func Stop(vmName string, forceKill bool, forceCleanup bool) error {
	// If the logger was already set, ignore this
	if !log.ConfigSet {
		log.SetFileLocation(HosterVmUtils.VM_AUDIT_LOG_LOCATION)
	}

	// Check if the VM exists block
	vms, err := HosterVmUtils.ListAllSimple()
	if err != nil {
		return err
	}
	vmFound := false
	for _, v := range vms {
		if v.VmName == vmName {
			vmFound = true
			break
		}
	}
	if !vmFound {
		return fmt.Errorf("%s: %s", HosterVmUtils.ERRTXT_VM_DOESNT_EXIST, vmName)
	}
	// EOF Check if the VM exists block

	// Check if the VM is running block
	vmsRunning, _ := HosterVmUtils.GetRunningVms()
	if !slices.Contains(vmsRunning, vmName) {
		return fmt.Errorf("%s: %s", HosterVmUtils.ERRTXT_VM_IS_STOPPED, vmName)
	}
	// EOF Check if the VM is running block

	err = SendShutdownSignal(vmName, forceKill, forceCleanup)
	if err != nil {
		return err
	}

	return nil
}

// Send a specific shutdown signal to your VM.
//
// It's cheap to call, but expects you to check if the VM exists and running in the first place.
func SendShutdownSignal(vmName string, forceKill bool, forceCleanup bool) error {
	// Pre-init vars
	vmPid := -1
	supervisorPid := -1
	if forceCleanup && !forceKill {
		return errors.New("cleanup parameter can only be used together with force kill parameter")
	}

	// If the logger was already set, ignore this
	if !log.ConfigSet {
		log.SetFileLocation(HosterVmUtils.VM_AUDIT_LOG_LOCATION)
	}

	pids, err := FreeBSDPgrep.Pgrep(vmName)
	// If the error happened, or if the list of PIDs is empty and we don't need to force kill - then throw an error.
	// If we do force kill, we don't want to stop here, so we'll just ignore the error.
	// TBD: Might be a good idea to log it as debug in the force kill mode. Needs a further discussion.
	if (err != nil || len(pids) < 1) && !forceKill {
		return errors.New("could not find the VM process specified (is the VM running?)")
	}

	// If there was no error, and PID list has more than one item - find a correct bhyve process send a respective kill signal to it
	if err == nil && len(pids) > 0 {
		reMatchVm := regexp.MustCompile(`bhyve:\s+` + vmName + `($|\s+)`)
		reMatchSupervisor := regexp.MustCompile(`/vm_supervisor_service for ` + vmName + `$`)
		for _, v := range pids {
			if reMatchVm.MatchString(v.ProcessCmd) {
				vmPid = v.ProcessId
				continue
			}
			if reMatchSupervisor.MatchString(v.ProcessCmd) {
				supervisorPid = v.ProcessId
				continue
			}
		}
	}
	// If the VM Pid was not found and force kill mode is off - return an error
	if vmPid < 0 && !forceKill {
		return errors.New("could not find the VM process specified (is the VM running?)")
	}
	// Perform bhyve process shutdown if the PID was found
	if vmPid >= 0 {
		if forceKill {
			FreeBSDKill.KillProcess(FreeBSDKill.KillSignalKILL, vmPid)
			message := fmt.Sprintf("Forceful SIGKILL signal has been sent to: %s; PID: %d", vmName, vmPid)
			log.Info(message)
		} else {
			FreeBSDKill.KillProcess(FreeBSDKill.KillSignalTERM, vmPid)
			message := fmt.Sprintf("Graceful SIGTERM signal has been sent to: %s; PID: %d", vmName, vmPid)
			log.Info(message)
		}
	}

	if forceKill || forceCleanup {
		err = HosterVmUtils.BhyveCtlForcePoweroff(vmName)
		if err != nil {
			log.Error(err.Error())
		}

		err := HosterVmUtils.BhyveCtlDestroy(vmName)
		if err != nil {
			log.Error(err.Error())
		}
	}

	if forceCleanup && supervisorPid >= 0 {
		// TBD: handle the graceful VM Supervisor shutdown
		// FreeBSDKill.KillProcess(FreeBSDKill.KillSignalTERM, supervisorPid)

		// Forceful VM Supervisor shutdown using -SIGKILL
		FreeBSDKill.KillProcess(FreeBSDKill.KillSignalKILL, supervisorPid)
		message := fmt.Sprintf("SIGKILL signal has been sent to the VM Supervisor; PID: %d", supervisorPid)
		log.Info(message)
	}

	// Clean-up the network interfaces, if they still exist, and log
	if forceCleanup {
		ifaces, err := HosterNetwork.VmNetworkCleanup(vmName)
		if err != nil {
			return err
		}
		for _, v := range ifaces {
			if v.Success {
				message := fmt.Sprintf("CMD SUCCESS: ifconfig %s destroy", v.IfaceName)
				log.Info(message)
			}
			if v.Failure {
				message := fmt.Sprintf("CMD ERR: ifconfig %s destroy", v.IfaceName)
				log.Info(message)
			}
		}
	}

	return nil
}
