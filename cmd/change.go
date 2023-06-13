package cmd

import (
	"errors"
	"hoster/emojlog"
	"log"

	"github.com/spf13/cobra"
	"golang.org/x/exp/slices"
)

var (
	changeCmd = &cobra.Command{
		Use:   "change",
		Short: "Change some of the settings",
		Long:  `Change some of the settings`,
		Run: func(cmd *cobra.Command, args []string) {
			err := checkInitFile()
			if err != nil {
				log.Fatal(err.Error())
			}
			cmd.Help()
		},
	}
)

var (
	changeParentVmName    string
	changeParentNewParent string
	changeParentCmd       = &cobra.Command{
		Use:   "parent",
		Short: "Change VM parent",
		Long:  `Change VM parent, in order to start this VM on a new host`,
		Run: func(cmd *cobra.Command, args []string) {
			err := checkInitFile()
			if err != nil {
				log.Fatal(err.Error())
			}
			err = replaceParent(changeParentVmName, changeParentNewParent)
			if err != nil {
				log.Fatal("could not set a new parent: " + err.Error())
			}
		},
	}
)

func replaceParent(vmName string, newParent string) error {
	if len(vmName) < 1 {
		return errors.New("you must provide a VM name")
	}
	if !slices.Contains(getAllVms(), vmName) {
		return errors.New("vm does not exist on this system")
	}
	if vmLiveCheck(vmName) {
		return errors.New("vm must be offline to perform this operation")
	}
	if len(newParent) < 1 {
		newParent = GetHostName()
	}

	vmFolder := getVmFolder(vmName)
	vmConfigVar := vmConfig(vmName)
	if vmConfigVar.ParentHost == newParent {
		emojlog.PrintLogMessage("No changes applied, because the old parent value is the same as a new parent value", emojlog.Debug)
		return nil
	}
	vmConfigVar.ParentHost = newParent

	err := vmConfigFileWriter(vmConfigVar, vmFolder+"/vm_config.json")
	if err != nil {
		return err
	}

	emojlog.PrintLogMessage("The VM parent has been changed to: "+newParent, emojlog.Changed)
	return nil
}
