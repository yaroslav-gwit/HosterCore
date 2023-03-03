package cmd

import (
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
			viewLog(args[0])
		},
	}
)

func viewLog(vmName string) {
	vmFolder := getVmFolder(vmName)
	tailCmd := exec.Command("tail", "-n", "35", "-f", vmFolder+"/vm_supervisor.log")
	tailCmd.Stdin = os.Stdin
	tailCmd.Stdout = os.Stdout
	tailCmd.Stderr = os.Stderr
	err := tailCmd.Run()
	if err != nil {
		log.Fatal("Can't open `tail -f` " + err.Error())
	}
}
