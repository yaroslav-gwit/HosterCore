//go:build freebsd
// +build freebsd

package cmd

import (
	"HosterCore/internal/pkg/emojlog"
	HosterVmUtils "HosterCore/internal/pkg/hoster/vm/utils"
	"encoding/json"
	"log"
	"os"

	"github.com/spf13/cobra"
)

var (
	jsonVmInfo       bool
	jsonPrettyVmInfo bool

	vmInfoCmd = &cobra.Command{
		Use:   "info [vmName]",
		Short: "Print out the VM Info",
		Long:  `Print out the VM Info.`,
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			checkInitFile()

			err := printVmInfo(args[0])
			if err != nil {
				emojlog.PrintLogMessage(err.Error(), emojlog.Error)
				os.Exit(1)
			}
		},
	}
)

func printVmInfo(vmName string) error {
	vmInfo, err := HosterVmUtils.InfoJsonApi(vmName)
	if err != nil {
		return err
	}

	if jsonPrettyVmInfo {
		jsonPretty, err := json.MarshalIndent(vmInfo, "", "   ")
		if err != nil {
			log.Fatal(err)
		}
		println(string(jsonPretty))
	} else {
		jsonOutput, err := json.Marshal(vmInfo)
		if err != nil {
			log.Fatal(err)
		}
		println(string(jsonOutput))
	}

	return nil
}
