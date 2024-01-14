package cmd

import (
	"HosterCore/internal/pkg/emojlog"
	HosterTables "HosterCore/internal/pkg/hoster/tables.go"
	"os"

	"github.com/spf13/cobra"
)

var (
	jailListCmdUnixStyle bool

	jailListCmd = &cobra.Command{
		Use:   "list",
		Short: "List all available Jails in a single table",
		Long:  `List all available Jails in a single table.`,
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			checkInitFile()

			err := HosterTables.GenerateJailsTable(jailListCmdUnixStyle)
			if err != nil {
				emojlog.PrintLogMessage(err.Error(), emojlog.Error)
				os.Exit(1)
			}
		},
	}
)
