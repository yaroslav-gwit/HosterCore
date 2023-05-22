package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"

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
	if len(newParent) < 1 {
		newParent = GetHostName()
	}

	if !slices.Contains(getAllVms(), vmName) {
		return errors.New("vm does not exist on this system")
	}
	vmFolder := getVmFolder(vmName)
	vmConfigVar := vmConfig(vmName)
	vmConfigVar.ParentHost = newParent

	jsonOutput, err := json.MarshalIndent(vmConfigVar, "", "   ")
	if err != nil {
		return err
	}
	fmt.Println(string(jsonOutput))

	// Open the file in write-only mode, truncating it if it already exists
	fmt.Println()
	fmt.Println(vmFolder + "vm_config.json")

	file, err := os.OpenFile(vmFolder+"vm_config.json", os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	// Write data to the file
	_, err = file.Write(jsonOutput)
	if err != nil {
		return err
	}

	return nil
}
