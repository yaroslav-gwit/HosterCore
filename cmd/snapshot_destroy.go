package cmd

import (
	"HosterCore/emojlog"
	"errors"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

var (
	snapshotDestroyCmd = &cobra.Command{
		Use:   "destroy [vmName] [snapshotName]",
		Short: "Destroy one of the VM's snapshots",
		Long:  `Destroy one of the VM's snapshots.`,
		Args:  cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			checkInitFile()
			err := ZfsSnapshotDestroy(args[0], args[1])
			if err != nil {
				emojlog.PrintLogMessage(err.Error(), emojlog.Error)
				os.Exit(1)
			}
		},
	}
)

// Destroys a snapshot for any given VM.
// Don't use in loops, because it performs some costly checks.
// Better come up with a sibling function that will ignore all of these, if the happy path is known beforehand.
func ZfsSnapshotDestroy(vmName string, snapshotName string) error {
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

	out, err := exec.Command("zfs", "destroy", snapshotName).CombinedOutput()
	if err != nil {
		return errors.New("something went wrong: " + string(out) + "; exit code: " + err.Error())
	}

	emojlog.PrintLogMessage("The snapshot has been destroyed: "+snapshotName, emojlog.Changed)
	return nil
}
