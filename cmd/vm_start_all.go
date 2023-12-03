package cmd

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
)

var (
	waitTime      int
	vmStartAllCmd = &cobra.Command{
		Use:   "start-all",
		Short: "Start all VMs deployed on this system",
		Long:  `Start all VMs deployed on this system`,
		Run: func(cmd *cobra.Command, args []string) {
			checkInitFile()
			vmStartAll(waitTime)
		},
	}
)

// Starts all production VMs, applying a wait time (in seconds) between issuing each `vm start` command
func vmStartAll(waitTime int) {
	allVms := getAllVms()
	offlineVms := []string{}

	for _, vm := range allVms {
		if VmLiveCheck(vm) {
			continue
		}

		vmConfigVar := vmConfig(vm)
		if IsVmInProduction(vmConfigVar.LiveStatus) {
			if vmConfigVar.ParentHost == GetHostName() {
				offlineVms = append(offlineVms, vm)
			}
		}
	}

	for i, v := range offlineVms {
		// Print out the output splitter
		if i == 0 {
			_ = 0
		} else {
			fmt.Println("  ───────────")
		}

		// Apply the sleep
		if i != 0 {
			time.Sleep(time.Second * time.Duration(waitTime))
		}

		// Start the VM
		VmStart(v, false, false, false)
	}
}

// Check if VM is in production using vmConfig.LiveStatus as input
func IsVmInProduction(s string) bool {
	if s == "production" || s == "prod" {
		return true
	}
	return false
}
