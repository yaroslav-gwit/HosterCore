package cmd

import (
	"github.com/spf13/cobra"
)

var (
	vmCmd = &cobra.Command{
		Use:   "vm",
		Short: "VM related operations",
		Long:  `VM related operations: VM deploy, VM stop, VM start, VM destroy, etc`,
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}
)
