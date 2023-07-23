package cmd

import (
	"log"

	"github.com/spf13/cobra"
)

var (
	snapshotCmd = &cobra.Command{
		Use:   "snapshot",
		Short: "Snapshot related commands",
		Long:  `Snapshot related commands.`,
		Run: func(cmd *cobra.Command, args []string) {
			err := checkInitFile()
			if err != nil {
				log.Fatal(err.Error())
			}

			cmd.Help()
		},
	}
)
