//go:build freebsd
// +build freebsd

package cmd

import (
	HosterVmUtils "HosterCore/internal/pkg/hoster/vm/utils"
	"errors"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/spf13/cobra"
)

var (
	vmSerialConsoleCmd = &cobra.Command{
		Use:   "serial-console [vmName]",
		Short: "Connect to VM's serial console",
		Long:  `Connect to VM's serial console (useful for connectivity troubleshooting)`,
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			checkInitFile()
			err := connectToSerialConsole(args[0])
			if err != nil {
				log.Fatal(err)
			}
		},
	}
)

func connectToSerialConsole(vmName string) error {
	vmInfo, err := HosterVmUtils.InfoJsonApi(vmName)
	if err != nil {
		return err
	}
	if !vmInfo.Running {
		return errors.New("vm is offline")
	}

	tmuxSessionList := exec.Command("tmux", "ls")
	stdout, stderr := tmuxSessionList.Output()

	if stderr != nil {
		if tmuxSessionList.ProcessState.ExitCode() == 1 {
			err := newTmuxSession(vmName)
			if err != nil {
				return errors.New("can't open VM console: " + err.Error())
			}
			return nil
		}
	}

	reTmuxSessionMatch := regexp.MustCompile(`^` + vmName + `:.*`)
	for _, v := range strings.Split(string(stdout), "\n") {
		v = strings.TrimSpace(v)
		if reTmuxSessionMatch.MatchString(v) {
			attachToTmuxSession(vmName)
			return nil
		}
	}

	newTmuxSession(vmName)
	return nil
}

func newTmuxSession(vmName string) error {
	tmuxCreate := exec.Command("tmux", "new-session", "-s", vmName, "/usr/bin/cu", "-l", "/dev/nmdm-"+vmName+"-1B")
	tmuxCreate.Stdin = os.Stdin
	tmuxCreate.Stdout = os.Stdout
	tmuxCreate.Stderr = os.Stderr

	err := tmuxCreate.Run()
	if err != nil {
		return errors.New("can't open VM console: " + err.Error())
	}
	return nil
}

func attachToTmuxSession(vmName string) error {
	tmuxAttach := exec.Command("tmux", "a", "-t", vmName)
	tmuxAttach.Stdin = os.Stdin
	tmuxAttach.Stdout = os.Stdout
	tmuxAttach.Stderr = os.Stderr

	err := tmuxAttach.Run()
	if err != nil {
		return errors.New("can't open VM console: " + err.Error())
	}
	return nil
}
