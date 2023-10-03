package cmd

import (
	"HosterCore/emojlog"
	"os"
	"time"

	"github.com/spf13/cobra"
)

var (
	jailStartAllCmd = &cobra.Command{
		Use:   "start-all",
		Short: "Start all available Jails on this system",
		Long:  `Start all available Jails on this system.`,
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			err := checkInitFile()
			if err != nil {
				emojlog.PrintLogMessage(err.Error(), emojlog.Error)
				os.Exit(1)
			}

			err = startAllJails(true)
			if err != nil {
				emojlog.PrintLogMessage(err.Error(), emojlog.Error)
				os.Exit(1)
			}
		},
	}
)

func startAllJails(consoleLogOutput bool) error {
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
		if jailOnline {
			continue
		}

		if i != 0 {
			time.Sleep(3 * time.Second)
		}

		err = jailStart(v, consoleLogOutput)
		if err != nil {
			return err
		}
	}

	return nil
}
