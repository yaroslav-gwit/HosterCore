package cmd

import (
	"HosterCore/pkg/emojlog"
	"fmt"
	"regexp"
	"time"

	"github.com/spf13/cobra"
)

var (
	vmCmd = &cobra.Command{
		Use:   "vm",
		Short: "VM related operations",
		Long:  `VM related operations: VM deploy, VM stop, VM start, VM destroy, etc`,
		Run: func(cmd *cobra.Command, args []string) {
			checkInitFile()
			cmd.Help()
		},
	}
)

var (
	vmUnlockAllCmd = &cobra.Command{
		Use:   "unlock-all",
		Short: "Unlock all HA locked VMs",
		Long:  `Unlock all HA locked VMs.`,
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			checkInitFile()
			UnlockAllVms()
			emojlog.PrintLogMessage("All VMs have now been unlocked", emojlog.Debug)
		},
	}
)

func LockAllVms() {
	allVms := GetAllVms()

	timeNow := time.Now().Format("2006-01-02_15-04-05")
	haLockedString := fmt.Sprintf("__HA_LOCKED_%s__", timeNow)

	for _, vm := range allVms {
		vmConfig := VmConfig(vm)
		if IsVmInProduction(vmConfig.LiveStatus) && vmConfig.ParentHost == GetHostName() {
			ReplaceParent(vm, haLockedString, true)
		}
	}
}

func UnlockAllVms() {
	allVms := GetAllVms()
	reHaLockedString := regexp.MustCompile(`__HA_LOCKED_.*`)

	for _, vm := range allVms {
		vmConfig := VmConfig(vm)
		if IsVmInProduction(vmConfig.LiveStatus) && reHaLockedString.MatchString(vmConfig.ParentHost) {
			ReplaceParent(vm, GetHostName(), true)
		}
	}
}
