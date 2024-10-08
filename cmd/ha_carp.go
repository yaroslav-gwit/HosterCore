//go:build freebsd
// +build freebsd

package cmd

import (
	CarpClient "HosterCore/internal/app/ha_carp/client"
	CarpUtils "HosterCore/internal/app/ha_carp/utils"
	"HosterCore/internal/pkg/emojlog"
	FreeBSDKill "HosterCore/internal/pkg/freebsd/kill"
	FreeBSDPgrep "HosterCore/internal/pkg/freebsd/pgrep"
	HosterLocations "HosterCore/internal/pkg/hoster/locations"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"syscall"

	"github.com/spf13/cobra"
)

var (
	carpHaCmd = &cobra.Command{
		Use:   "ha",
		Short: "HA/CARP related operations",
		Long:  `HA/CARP related operations.`,
		Run: func(cmd *cobra.Command, args []string) {
			checkInitFile()
			cmd.Help()
		},
	}
)

var (
	haStartCmd = &cobra.Command{
		Use:   "start",
		Short: "Start HA Mode",
		Long:  `Start HA Mode.`,
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			checkInitFile()

			err := startHaCarp()
			if err != nil {
				emojlog.PrintErrorMessage("service could not be started -> " + err.Error())
			} else {
				emojlog.PrintChangedMessage("HA service has been started")
			}

			os.Exit(0)
		},
	}
)

var (
	haStatusCmd = &cobra.Command{
		Use:   "status",
		Short: "HA Mode Status",
		Long:  `HA Mode Status.`,
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			checkInitFile()

			pid, _ := FreeBSDPgrep.FindHaCarp()
			if pid > 0 {
				logVal := fmt.Sprintf("HA service is running with PID: %d", pid)
				emojlog.PrintInfoMessage(logVal)
			} else {
				logVal := "HA service is not running"
				emojlog.PrintWarningMessage(logVal)
			}

			os.Exit(0)
		},
	}
)

var (
	haStopCmd = &cobra.Command{
		Use:   "stop",
		Short: "Stop HA Mode",
		Long:  `Stop HA Mode.`,
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			checkInitFile()

			pid, _ := FreeBSDPgrep.FindHaCarp()
			if pid < 1 {
				emojlog.PrintWarningMessage("HA service is not running")
				os.Exit(1)
			}

			err := FreeBSDKill.KillProcess(FreeBSDKill.KillSignalTERM, pid)
			if err != nil {
				emojlog.PrintErrorMessage("service could not be stopped -> " + err.Error())
			} else {
				emojlog.PrintChangedMessage("HA service has been stopped")
			}
		},
	}
)

var (
	haInfoCmd = &cobra.Command{
		Use:   "info",
		Short: "Get real-time HA Mode information",
		Long:  "Get real-time HA Mode information.",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			checkInitFile()

			out, err := CarpClient.GetHaStatus()
			if err != nil {
				emojlog.PrintErrorMessage("could not read HA status -> " + err.Error())
				os.Exit(1)
			}

			jsonOut, err := json.MarshalIndent(out, "", "  ")
			if err != nil {
				emojlog.PrintErrorMessage("could not convert HA status into JSON -> " + err.Error())
				os.Exit(1)
			}
			fmt.Println(string(jsonOut))

			os.Exit(0)
		},
	}
)

var (
	haShowLogCmd = &cobra.Command{
		Use:   "show-log",
		Short: "Get real-time HA Mode logs",
		Long:  "Get real-time HA Mode logs.",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			checkInitFile()

			tailCmd := exec.Command("tail", "-35", "-f", CarpUtils.LOG_FILE)
			tailCmd.Stdin = os.Stdin
			tailCmd.Stdout = os.Stdout
			tailCmd.Stderr = os.Stderr

			err := tailCmd.Run()
			if err != nil {
				emojlog.PrintErrorMessage("could not show log -> " + err.Error())
				os.Exit(1)
			}

			os.Exit(0)
		},
	}
)

func startHaCarp() error {
	pid, _ := FreeBSDPgrep.FindHaCarp()
	if pid > 0 {
		return fmt.Errorf("HA service is already running with PID: %d", pid)
	}

	binLoc, err := HosterLocations.LocateBinary("ha_carp")
	if err != nil {
		return err
	}

	command := exec.Command(binLoc)
	command.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	err = command.Start()
	if err != nil {
		return err
	}

	return nil
}
