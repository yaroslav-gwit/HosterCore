package cmd

import (
	"github.com/spf13/cobra"
)

var (
	snapshotAllCmdType    string
	snapshotsAllCmdToKeep int

	snapshotAllCmd = &cobra.Command{
		Use:   "all",
		Short: "Snapshot all running VMs on this system",
		Long:  `Snapshot all running VMs on this system.`,
		Run: func(cmd *cobra.Command, args []string) {
			checkInitFile()
			snapshotAllRunningVms()
		},
	}
)

func snapshotAllRunningVms() {
	for _, vm := range getAllVms() {
		if VmLiveCheck(vm) {
			VmZfsSnapshot(vm, snapshotAllCmdType, snapshotsAllCmdToKeep)
		}
	}
}
