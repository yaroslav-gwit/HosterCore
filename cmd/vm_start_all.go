package cmd

import (
	"log"
	"time"

	"github.com/spf13/cobra"
)

var (
	waitTime int
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
	if waitTime < 1 {
		sleepTime := 5
		for i, vm := range getAllVms() {
			vmConfigVar := vmConfig(vm)
			if vmConfigVar.ParentHost != GetHostName() {
				continue
			}
			if vmConfigVar.LiveStatus == "production" || vmConfigVar.LiveStatus == "prod" {
				_ = 0
			} else {
				continue
			}
			if i > 0 {
				time.Sleep(time.Second * time.Duration(sleepTime))
			}
			if sleepTime < 30 {
				sleepTime = sleepTime + 1
			}
			vmStart(vm)
		}
	} else {
		for i, vm := range getAllVms() {
			vmConfigVar := vmConfig(vm)
			if vmConfigVar.ParentHost != GetHostName() {
				continue
			}
			if vmConfigVar.LiveStatus == "production" || vmConfigVar.LiveStatus == "prod" {
				_ = 0
			} else {
				continue
			}
			if i > 0 {
				time.Sleep(time.Second * time.Duration(waitTime))
			}
			vmStart(vm)
		}
	}
}
