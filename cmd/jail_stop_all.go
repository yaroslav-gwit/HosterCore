package cmd

import (
	"HosterCore/emojlog"
	"os"
	"time"

	"github.com/spf13/cobra"
)

var (
	jailStopAllCmd = &cobra.Command{
		Use:   "stop-all",
		Short: "Stop all online Jails on this system",
		Long:  `Stop all online Jails on this system.`,

		Args: cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			err := checkInitFile()
			if err != nil {
				emojlog.PrintLogMessage(err.Error(), emojlog.Error)
				os.Exit(1)
			}

			err = stopAllJails(true)
			if err != nil {
				emojlog.PrintLogMessage(err.Error(), emojlog.Error)
				os.Exit(1)
			}
		},
	}
)

func stopAllJails(consoleLogOutput bool) error {
	jailList, err := getAllJailsList()
	if err != nil {
		return err
	}

	for i, v := range jailList {
		jailConfig, err := getJailConfig(v, true)
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

		if i != 0 {
			time.Sleep(3 * time.Second)
		}

		err = jailStop(v, consoleLogOutput)
		if err != nil {
			return err
		}
	}

	return nil
}
