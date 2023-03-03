package cmd

import (
	"github.com/spf13/cobra"
)

var (
	snapshotAllType    string
	snapshotsAllToKeep int

	vmZfsSnapshotAllCmd = &cobra.Command{
		Use:   "snapshot-all",
		Short: "Snapshot all running VMs on this system",
		Long:  `Snapshot all running VMs on this system`,
		Run: func(cmd *cobra.Command, args []string) {
			snapshotAllRunningVms()
		},
	}
)

func snapshotAllRunningVms() {
	for _, vm := range getAllVms() {
		if vmLiveCheck(vm) {
			vmZfsSnapshot(vm, snapshotAllType, snapshotsAllToKeep)
		}
	}
}
