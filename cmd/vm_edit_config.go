package cmd

import (
	"log"
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
			manuallyEditConfig(args[0])
		},
	}
)

func manuallyEditConfig(vmName string) {
	vmFolder := getVmFolder(vmName)
	textEditor := os.Getenv("EDITOR")

	if len(textEditor) < 1 {
		textEditor = "micro"
	}

	tailCmd := exec.Command(textEditor, vmFolder+"/vm_config.json")
	tailCmd.Stdin = os.Stdin
	tailCmd.Stdout = os.Stdout
	tailCmd.Stderr = os.Stderr

	err := tailCmd.Run()
	if err != nil {
		log.Fatal("Can't open your editor: " + err.Error())
	}
}
