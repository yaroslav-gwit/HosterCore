package cmd

// var (
// 	snapshotAllCmdType    string
// 	snapshotsAllCmdToKeep int

// 	snapshotAllCmd = &cobra.Command{
// 		Use:   "all",
// 		Short: "Snapshot all running VMs on this system",
// 		Long:  `Snapshot all running VMs on this system.`,
// 		Run: func(cmd *cobra.Command, args []string) {
// 			checkInitFile()
// 			snapshotAllRunningVms()
// 		},
// 	}
// )

// func snapshotAllRunningVms() {
// 	vms, _ := HosterVmUtils.ListJsonApi()

// 	for _, v := range vms {
// 		if v.Running {
// 			VmZfsSnapshot(v.Name, snapshotAllCmdType, snapshotsAllCmdToKeep)
// 		}
// 	}
// }
