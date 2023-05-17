package cmd

import (
	"log"
	"time"

	"github.com/spf13/cobra"
)

var (
	forceStopAll bool
	vmStopAllCmd = &cobra.Command{
		Use:   "stop-all",
		Short: "Stop all VMs deployed on this system",
		Long:  `Stop all VMs deployed on this system`,
		Run: func(cmd *cobra.Command, args []string) {
			err := checkInitFile()
			if err != nil {
				log.Fatal(err.Error())
			}

			vmStopAll(forceStopAll)
		},
	}
)

func vmStopAll(forceStopAll bool) {
	sleepTime := 3
	for i, vm := range getAllVms() {
		if i != 0 {
			time.Sleep(time.Second * time.Duration(sleepTime))
		}
		vmStop(vm, false)
	}
}
