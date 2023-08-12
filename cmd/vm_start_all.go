package cmd

import (
	"log"
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
			err := checkInitFile()
			if err != nil {
				log.Fatal(err.Error())
			}

			vmStartAll(waitTime)
		},
	}
)

// Starts all production VMs, applying a wait time (in seconds) between issuing each `vm start` command
func vmStartAll(waitTime int) {
	allVms := getAllVms()
	iteration := 0
	for _, vm := range allVms {
		vmConfigVar := vmConfig(vm)
		if vmConfigVar.ParentHost != GetHostName() {
			continue
		}
		if !VmIsInProduction(vmConfigVar.LiveStatus) {
			continue
		}
		iteration = iteration + 1
		if iteration != 1 {
			time.Sleep(time.Second * time.Duration(waitTime))
		}
		VmStart(vm)
	}
}

// Check if VM is in production using vmConfig.LiveStatus as input
func VmIsInProduction(s string) bool {
	if s == "production" || s == "prod" {
		return true
	}
	return false
}
