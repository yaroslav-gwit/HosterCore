//go:build freebsd
// +build freebsd

package cmd

import (
	"HosterCore/internal/pkg/emojlog"
	HosterVmUtils "HosterCore/internal/pkg/hoster/vm/utils"
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

var (
	vmEditConfigCmd = &cobra.Command{
		Use:   "edit-config [vmName]",
		Short: "Edit VM's config manually using your favorite text editor",
		Long:  `Edit VM's config manually using your favorite text editor`,
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			checkInitFile()

			err := manuallyEditConfig(args[0])
			if err != nil {
				emojlog.PrintLogMessage(err.Error(), emojlog.Error)
				os.Exit(1)
			}
		},
	}
)

func manuallyEditConfig(vmName string) error {
	vmInfo, err := HosterVmUtils.InfoJsonApi(vmName)
	if err != nil {
		return fmt.Errorf("can't open your editor: %s", err.Error())
	}

	vmFolder := vmInfo.Simple.Mountpoint + "/" + vmInfo.Name
	textEditor := os.Getenv("EDITOR")

	if len(textEditor) < 1 {
		textEditor = "vi"
	}

	tailCmd := exec.Command(textEditor, vmFolder+"/vm_config.json")
	tailCmd.Stdin = os.Stdin
	tailCmd.Stdout = os.Stdout
	tailCmd.Stderr = os.Stderr

	err = tailCmd.Run()
	if err != nil {
		return fmt.Errorf("can't open your editor: %s", err.Error())
	}

	return nil
}
