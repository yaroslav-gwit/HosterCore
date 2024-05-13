//go:build freebsd
// +build freebsd

package cmd

import (
	HosterVmUtils "HosterCore/internal/pkg/hoster/vm/utils"
	"log"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

var (
	vmShowLogCmd = &cobra.Command{
		Use:   "show-log [vmName]",
		Short: "Show log in real time using `tail -f`",
		Long:  `Show log in real time using "tail -f"`,
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			checkInitFile()
			viewLog(args[0])
		},
	}
)

func viewLog(vmName string) {
	vmFolder := ""
	vms, err := HosterVmUtils.ListAllSimple()
	if err != nil {
		log.Fatal("Can't open `tail -f` " + err.Error())
	}

	for _, v := range vms {
		if v.VmName == vmName {
			vmFolder = v.Mountpoint + "/" + vmName
		}
	}

	tailCmd := exec.Command("tail", "-n", "35", "-f", vmFolder+"/vm_supervisor.log")
	tailCmd.Stdin = os.Stdin
	tailCmd.Stdout = os.Stdout
	tailCmd.Stderr = os.Stderr

	err = tailCmd.Run()
	if err != nil {
		log.Fatal("Can't open `tail -f` " + err.Error())
	}
}
