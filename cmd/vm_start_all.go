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

func vmStartAll(waitTime int) {
	allVms := getAllVms()
	if waitTime < 1 {
		sleepTime := 5
		for i, vm := range allVms {
			vmConfigVar := vmConfig(vm)
			if vmConfigVar.ParentHost != GetHostName() {
				continue
			}
			if !vmIsInProduction(vmConfigVar.LiveStatus) {
				continue
			}
			if i == len(allVms)-1 {
				_ = 0
			} else if sleepTime < 30 {
				sleepTime = sleepTime + 1
			}
			vmStart(vm)
		}
	} else {
		for i, vm := range allVms {
			vmConfigVar := vmConfig(vm)
			if vmConfigVar.ParentHost != GetHostName() {
				continue
			}
			if !vmIsInProduction(vmConfigVar.LiveStatus) {
				continue
			}
			if i == len(allVms)-1 {
				_ = 0
			} else if i > 0 {
				time.Sleep(time.Second * time.Duration(waitTime))
			}
			vmStart(vm)
		}
	}
}

// Check if VM is in production using vmConfig.LiveStatus as input
func vmIsInProduction(s string) bool {
	if s == "production" || s == "prod" {
		return true
	}
	return false
}
