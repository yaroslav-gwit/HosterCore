package cmd

import (
	HosterCliJson "HosterCore/internal/pkg/hoster/cli_json"
	HosterTables "HosterCore/internal/pkg/hoster/cli_tables"

	"github.com/spf13/cobra"
)

var (
	jsonHostInfoOutput       bool
	jsonPrettyHostInfoOutput bool

	hostCmd = &cobra.Command{
		Use:   "host",
		Short: "Host related operations",
		Long:  `Host related operations.`,
		Run: func(cmd *cobra.Command, args []string) {
			checkInitFile()

			if jsonHostInfoOutput || jsonPrettyHostInfoOutput {
				HosterCliJson.GenerateHostInfoJson(jsonPrettyHostInfoOutput)
			} else {
				HosterTables.GenerateHostInfoTable(false)
			}
		},
	}
)

// Console color outputs
const LIGHT_RED = "\033[38;5;203m"
const LIGHT_GREEN = "\033[0;92m"
const LIGHT_YELLOW = "\033[0;93m"
const NC = "\033[0m"
