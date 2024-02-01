package cmd

import (
	HosterVm "HosterCore/internal/pkg/hoster/vm"
	"time"

	"github.com/spf13/cobra"
)

var (
	forceStopAll        bool
	vmStopAllCmdCleanUp bool
	vmStopAllCmd        = &cobra.Command{
		Use:   "stop-all",
		Short: "Stop all VMs deployed on this system",
		Long:  `Stop all VMs deployed on this system`,
		Run: func(cmd *cobra.Command, args []string) {
			checkInitFile()
			VmStopAll(forceStopAll, vmStopAllCmdCleanUp)
		},
	}
)

func VmStopAll(forceStopAll bool, cleanup bool) {
	sleepTime := 500
	for i, vm := range getAllVms() {
		if i != 0 {
			time.Sleep(time.Millisecond * time.Duration(sleepTime))
		}
		if forceStopAll {
			// VmStop(vm, true, vmStopAllCmdCleanUp)
			HosterVm.Stop(vm, true, vmStopAllCmdCleanUp)
		} else {
			// VmStop(vm, false, false)
			HosterVm.Stop(vm, false, false)
		}
	}
}
