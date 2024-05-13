//go:build freebsd
// +build freebsd

package cmd

import (
	"HosterCore/internal/pkg/emojlog"
	HosterVm "HosterCore/internal/pkg/hoster/vm"
	"os"

	"github.com/spf13/cobra"
)

var (
	vmSetCmd = &cobra.Command{
		Use:   "set",
		Short: "Change or set a particular VM config option",
		Long:  "Change or set a particular VM config option.",
		Run: func(cmd *cobra.Command, args []string) {
			checkInitFile()
			cmd.Help()
		},
	}
)

var (
	vmSetNewParent string

	vmSetCmdParent = &cobra.Command{
		Use:   "parent [vmName]",
		Short: "Change VM's parent",
		Long:  `Change VM's parent, in order to start it on a new host.`,
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			checkInitFile()

			err := HosterVm.ChangeParent(args[0], vmSetNewParent, false)
			if err != nil {
				emojlog.PrintLogMessage(err.Error(), emojlog.Error)
				os.Exit(1)
			}
		},
	}
)
