//go:build freebsd
// +build freebsd

package cmd

import (
	"HosterCore/internal/pkg/emojlog"
	FreeBSDKill "HosterCore/internal/pkg/freebsd/kill"
	FreeBSDPgrep "HosterCore/internal/pkg/freebsd/pgrep"
	HosterLocations "HosterCore/internal/pkg/hoster/locations"
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
