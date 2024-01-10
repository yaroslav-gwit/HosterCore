package cmd

import (
	"HosterCore/emojlog"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	jailStopAllCmd = &cobra.Command{
		Use:   "stop-all",
		Short: "Stop all online Jails on this system",
		Long:  `Stop all online Jails on this system.`,

		Args: cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			checkInitFile()

			err := stopAllJails(true)
			if err != nil {
				emojlog.PrintLogMessage(err.Error(), emojlog.Error)
				os.Exit(1)
			}
		},
	}
)

func stopAllJails(consoleLogOutput bool) error {
	jailList, err := GetAllJailsList()
	if err != nil {
		return err
	}

	startId := 0
	for _, v := range jailList {
		jailConfig, err := GetJailConfig(v, true)
		if err != nil {
			return err
		}
		jailOnline, err := checkJailOnline(jailConfig)
		if err != nil {
			return err
		}
		if !jailOnline {
			continue
		}

		// Print out the output splitter
		if startId == 0 {
			_ = 0
		} else {
			fmt.Println("  ───────────")
		}

		err = jailStop(v, consoleLogOutput)
		if err != nil {
			return err
		}

		startId += 1
	}

	return nil
}
