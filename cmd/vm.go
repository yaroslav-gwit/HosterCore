package cmd

import (
	"log"

	"github.com/spf13/cobra"
)

var (
	vmCmd = &cobra.Command{
		Use:   "vm",
		Short: "VM related operations",
		Long:  `VM related operations: VM deploy, VM stop, VM start, VM destroy, etc`,
		Run: func(cmd *cobra.Command, args []string) {
			err := checkInitFile()
			if err != nil {
				log.Fatal(err.Error())
			}

			cmd.Help()
		},
	}
)

func LockAllVms() {
	allVms := GetAllVms()
	for _, vm := range allVms {
		vmConfig := VmConfig(vm)
		if VmIsInProduction(vmConfig.LiveStatus) && vmConfig.ParentHost == GetHostName() {
			ReplaceParent(vm, "__HA_LOCKED__", true)
		}
	}
}
