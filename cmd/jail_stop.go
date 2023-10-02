package cmd

import (
	"HosterCore/emojlog"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
)

var (
	jailStopCmd = &cobra.Command{
		Use:   "stop [jailName]",
		Short: "Stop a specific Jail",
		Long:  `Stop a specific Jail using it's name`,
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			err := checkInitFile()
			if err != nil {
				emojlog.PrintLogMessage(err.Error(), emojlog.Error)
				os.Exit(1)
			}

			err = jailStop(args[0], true)
			if err != nil {
				emojlog.PrintLogMessage(err.Error(), emojlog.Error)
				os.Exit(1)
			}
		},
	}
)

func jailStop(jailName string, logActions bool) error {
	jailConfig, err := getJailConfig(jailName, false)
	if err != nil {
		return err
	}

	if logActions {
		emojlog.PrintLogMessage("Stopping the Jail: "+jailName, emojlog.Info)
	}

	// jailStopCommand := []string{"jexec", jailName}
	// splitAtSpace := regexp.MustCompile(`\s+`)
	// jailStopCommand = append(jailStopCommand, splitAtSpace.Split(jailConfig.ShutdownScript, -1)...)
	// if logActions {
	// 	emojlog.PrintLogMessage("Executing the Jail shutdown script: "+strings.Join(jailStopCommand, " "), emojlog.Debug)
	// }
	// jailStopOutput, err := exec.Command(jailStopCommand[0], jailStopCommand[1:]...).CombinedOutput()
	// if err != nil {
	// 	errorValue := errors.New("FATAL: " + string(jailStopOutput) + "; " + err.Error())
	// 	return errorValue
	// }

	// jailRemoveCommand := []string{"jail", "-r", jailName}
	// if logActions {
	// 	emojlog.PrintLogMessage("Cleaning up the Jail state: "+strings.Join(jailRemoveCommand, " "), emojlog.Debug)
	// }
	// jailRemoveOutput, err := exec.Command("jail", "-r", jailName).CombinedOutput()
	// if err != nil {
	// 	errorValue := errors.New("FATAL: " + string(jailRemoveOutput) + "; " + err.Error())
	// 	return errorValue
	// }

	_ = os.Remove("/etc/jail.conf")
	input, err := os.ReadFile(jailConfig.JailFolder + "jail_temp_runtime.conf")
	if err != nil {
		return err
	}
	err = os.WriteFile("/etc/jail.conf", input, 0644)
	if err != nil {
		return err
	}

	emojlog.PrintLogMessage(fmt.Sprintf("Stopping a Jail: %s", jailName), emojlog.Debug)
	out, err := exec.Command("service", "jail", "onestop", jailName).CombinedOutput()
	if err != nil {
		errorValue := "FATAL: " + strings.TrimSpace(string(out)) + "; " + err.Error()
		return fmt.Errorf("%s", errorValue)
	}

	ifconfigIpRemoveCommand := []string{"ifconfig", "vm-" + jailConfig.Network, jailConfig.IPAddress, "delete"}
	if logActions {
		emojlog.PrintLogMessage("Cleaning up the Jail IPs: "+strings.Join(ifconfigIpRemoveCommand, " "), emojlog.Debug)
	}
	ifconfigIpRemoveOutput, err := exec.Command("ifconfig", "vm-"+jailConfig.Network, jailConfig.IPAddress, "delete").CombinedOutput()
	if err != nil {
		errorValue := errors.New("FATAL: " + string(ifconfigIpRemoveOutput) + "; " + err.Error())
		return errorValue
	}

	clearJailUptimeStateFile(jailName)
	_ = os.Remove("/etc/jail.conf")

	if logActions {
		emojlog.PrintLogMessage("Cleared Jail uptime state file", emojlog.Changed)
		emojlog.PrintLogMessage("The Jail has been stopped: "+jailName, emojlog.Changed)
	}

	return nil
}
