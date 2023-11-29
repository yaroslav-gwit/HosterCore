package cmd

import (
	"HosterCore/emojlog"
	"errors"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

var (
	snapshotRollbackForceStop  = false
	snapshotRollbackForceStart = false

	snapshotRollbackCmd = &cobra.Command{
		Use:   "rollback [vmName] [snapshotName]",
		Short: "Rollback the VM to one of it's previous states",
		Long:  `Rollback the VM to one of it's previous states.`,
		Args:  cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			checkInitFile()
			err := ZfsSnapshotRollback(args[0], args[1], snapshotRollbackForceStop, snapshotRollbackForceStart)
			if err != nil {
				emojlog.PrintLogMessage(err.Error(), emojlog.Error)
				os.Exit(1)
			}
		},
	}
)

// Rolls back the VM to a previous ZFS snapshot, and destroys any newer snapshot, that has been taken after it.
// Can take "force" parameter in, that will stop the VM automatically, using "stop --force".
func ZfsSnapshotRollback(vmName string, snapshotName string, forceStop bool, forceStart bool) error {
	allVms := getAllVms()
	vmFound := false
	for _, v := range allVms {
		if v == vmName {
			vmFound = true
			break
		}
	}
	if !vmFound {
		return errors.New(VM_DOESNT_EXIST_STRING)
	}

	if forceStop {
		err := VmStop(vmName, forceStop, true)
		if err != nil {
			return err
		}
	} else if VmLiveCheck(vmName) {
		return errors.New("VM is online, please make sure VM is turned off before trying again")
	}

	snapInfoList, err := GetSnapshotInfo(vmName, false)
	if err != nil {
		return err
	}
	snapFound := false
	for _, v := range snapInfoList {
		if v.Name == snapshotName {
			snapFound = true
		}
	}
	if !snapFound {
		return errors.New("snapshot specified doesn't exist")
	}

	out, err := exec.Command("zfs", "rollback", "-r", snapshotName).CombinedOutput()
	if err != nil {
		return errors.New("something went wrong: " + string(out) + "; exit code: " + err.Error())
	}

	emojlog.PrintLogMessage("VM has been rolled back to: "+snapshotName, emojlog.Changed)

	if forceStart {
		err := VmStart(vmName, false, false, false)
		if err != nil {
			return err
		}
	}

	return nil
}
