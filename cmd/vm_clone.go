package cmd

import (
	"HosterCore/internal/pkg/emojlog"
	"errors"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

var (
	vmCloneCmd = &cobra.Command{
		Use:   "clone [existingVmName] [newVmName]",
		Short: "Use OpenZFS to clone your VM",
		Long:  `Use OpenZFS to clone your VM. You'll need to run "hoster vm cireset [newVmName]" in case the new VM has to be used as a separate machine.`,
		Args:  cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			checkInitFile()
			err := executeVmClone(args[0], args[1])
			if err != nil {
				emojlog.PrintLogMessage(err.Error(), emojlog.Error)
				os.Exit(1)
			}
		},
	}
)

// Executes `zfs clone` in the shell, returns an error if something has gone wrong
func executeVmClone(existingVmName string, newVmName string) error {
	// To list all datasets which were used as clones, run this (for later use):
	// zfs list -o name,origin

	// Find a current VM dataset that we can work with
	currentDataset, err := getVmDataset(existingVmName)
	if err != nil {
		return err
	}

	// Generate a new snapshot name, store it as a variable and execute zfs snapshot
	timeNow := time.Now().Format("2006-01-02_15-04-05.000")
	snapshotCloneName := currentDataset + "@clone_" + newVmName + "_" + timeNow
	out, err := exec.Command("zfs", "snapshot", snapshotCloneName).CombinedOutput()
	if err != nil {
		return errors.New(strings.TrimSpace(string(out)) + "; " + err.Error())
	}

	// Generate new dataset name and execute zfs clone
	newDataset := strings.TrimSuffix(currentDataset, "/"+existingVmName)
	newDataset = newDataset + "/" + newVmName
	outClone, err := exec.Command("zfs", "clone", snapshotCloneName, newDataset).CombinedOutput()
	if err != nil {
		return errors.New(strings.TrimSpace(string(outClone)) + "; " + err.Error())
	}

	// Reload the internal DNS server
	err = ReloadDnsServer()
	if err != nil {
		return err
	}

	// Report the status back to user
	emojlog.PrintLogMessage("Your VM has been cloned successfully", emojlog.Info)
	emojlog.PrintLogMessage("You might want to run `hoster vm cireset "+newVmName+"` to use it as an independent machine", emojlog.Info)

	return nil
}
