package cmd

import (
	"time"

	"github.com/spf13/cobra"
)

var (
	vmStartAllCmd = &cobra.Command{
		Use:   "start-all",
		Short: "Start all VMs deployed on this system",
		Long:  `Start all VMs deployed on this system`,
		Run: func(cmd *cobra.Command, args []string) {
			vmStartAll()
		},
	}
)

func vmStartAll() {
	sleepTime := 5
	for i, vm := range getAllVms() {
		vmConfigVar := vmConfig(vm)
		if vmConfigVar.ParentHost != GetHostName() {
			continue
		} else if vmConfigVar.LiveStatus == "production" || vmConfigVar.LiveStatus == "prod" {
			if i != 0 {
				time.Sleep(time.Second * time.Duration(sleepTime))
			}
			vmStart(vm)
			if sleepTime < 30 {
				sleepTime = sleepTime + 1
			}
		} else {
			continue
		}
	}
}
