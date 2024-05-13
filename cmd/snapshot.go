//go:build freebsd
// +build freebsd

package cmd

import (
	"github.com/spf13/cobra"
)

var (
	snapshotCmd = &cobra.Command{
		Use:   "snapshot",
		Short: "Snapshot related commands",
		Long:  `Snapshot related commands.`,
		Run: func(cmd *cobra.Command, args []string) {
			checkInitFile()
			cmd.Help()
		},
	}
)
