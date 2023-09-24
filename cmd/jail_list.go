package cmd

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"
)

var (
	jailListCmd = &cobra.Command{
		Use:   "list",
		Short: "List all available Jails in a single table",
		Long:  `List all available Jails in a single table.`,
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			err := checkInitFile()
			if err != nil {
				log.Fatal(err.Error())
			}
			fmt.Println(getRunningJails())
			fmt.Println()
			fmt.Println()
			fmt.Println(getAllJailsList())
		},
	}
)
