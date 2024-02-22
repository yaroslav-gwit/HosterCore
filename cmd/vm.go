package cmd

import (
	"HosterCore/internal/pkg/emojlog"
	HosterTables "HosterCore/internal/pkg/hoster/cli_tables"
	HosterVm "HosterCore/internal/pkg/hoster/vm"
	"fmt"
	"os"
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
	vmStartCmdRestoreVmState bool
	vmStartCmdWaitForVnc     bool
	vmStartCmdDebug          bool

	vmStartCmd = &cobra.Command{
		Use:   "start [vmName]",
		Short: "Start a particular VM using it's name",
		Long:  `Start a particular VM using it's name`,
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			checkInitFile()

			err := HosterVm.Start(args[0], vmStartCmdWaitForVnc, vmStartCmdDebug)
			if err != nil {
				emojlog.PrintLogMessage(err.Error(), emojlog.Error)
				os.Exit(1)
			}
		},
	}
)

var (
	vmStopCmdForceStop bool
	vmStopCmdCleanUp   bool
	vmStopCmd          = &cobra.Command{
		Use:   "stop [vmName]",
		Short: "Stop a particular VM using it's name",
		Long:  `Stop a particular VM using it's name`,
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			checkInitFile()

			err := HosterVm.Stop(args[0], vmStopCmdForceStop, vmStopCmdCleanUp)
			if err != nil {
				emojlog.PrintLogMessage(err.Error(), emojlog.Error)
				os.Exit(1)
			}
		},
	}
)

var (
	vmCloneSnapshot string

	vmCloneCmd = &cobra.Command{
		Use:   "clone [existingVmName] [newVmName]",
		Short: "Use OpenZFS to clone your VM",
		Long:  `Use OpenZFS to clone your VM. You'll need to run "hoster vm cireset [newVmName]" in case the new VM has to be used as a separate machine.`,
		Args:  cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			checkInitFile()

			err := HosterVm.Clone(args[0], args[1], vmCloneSnapshot)
			if err != nil {
				emojlog.PrintLogMessage(err.Error(), emojlog.Error)
				os.Exit(1)
			}
		},
	}
)

var (
	vmDestroyCmd = &cobra.Command{
		Use:   "destroy",
		Short: "Destroy the VM",
		Long:  `Destroy the VM and it's parent snapshot (uses zfs destroy)`,
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			checkInitFile()

			err := HosterVm.Destroy(args[0])
			if err != nil {
				emojlog.PrintLogMessage("Could not destroy the VM: "+err.Error(), emojlog.Error)
				os.Exit(1)
			}

			err = ReloadDnsServer()
			if err != nil {
				emojlog.PrintLogMessage("Could not reload the DNS server: "+err.Error(), emojlog.Error)
				os.Exit(1)
			}
		},
	}
)

var (
	jsonOutputVm       bool
	jsonPrettyOutputVm bool
	tableUnixOutputVm  bool

	vmListCmd = &cobra.Command{
		Use:   "list",
		Short: "VM list",
		Long:  `VM list in the form of tables, json, or json pretty`,
		Run: func(cmd *cobra.Command, args []string) {
			checkInitFile()

			err := HosterTables.GenerateVMsTable(tableUnixOutputVm)
			if err != nil {
				emojlog.PrintLogMessage(err.Error(), emojlog.Error)
				os.Exit(1)
			}
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
