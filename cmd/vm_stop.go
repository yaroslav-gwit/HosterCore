package cmd

import (
	"errors"
	"hoster/emojlog"
	"log"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"golang.org/x/exp/slices"
)

var (
	vmStopCmd = &cobra.Command{
		Use:   "stop [vmName]",
		Short: "Stop a particular VM using it's name",
		Long:  `Stop a particular VM using it's name`,
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			err := vmStop(args[0])
			if err != nil {
				log.Fatal(err)
			}
		},
	}
)

func vmStop(vmName string) error {
	allVms := getAllVms()
	if !slices.Contains(allVms, vmName) {
		return errors.New("VM is not found on this system")
	} else if !vmLiveCheck(vmName) {
		return errors.New("VM is already stopped")
	}

	StopBhyveProcess(vmName)
	vmSupervisorCleanup(vmName)
	StopBhyveProcess(vmName)
	NetworkCleanup(vmName, false)
	BhyvectlDestroy(vmName, false)

	return nil
}

func StopBhyveProcess(vmName string) {
	emojlog.PrintLogMessage("Stopping the VM: "+vmName, emojlog.Info)
	cmd := exec.Command("pgrep", "-lf", vmName)
	stdout, stderr := cmd.Output()
	if stderr != nil {
		if cmd.ProcessState.ExitCode() == 1 {
			_ = 0
		} else {
			emojlog.PrintLogMessage("pgrep exited with an error: "+stderr.Error(), emojlog.Error)
		}
	}

	processId := ""
	reMatchVm, _ := regexp.Compile(`.*bhyve:\s` + vmName)
	for _, v := range strings.Split(string(stdout), "\n") {
		if reMatchVm.MatchString(v) {
			processId = strings.TrimSpace(strings.Split(v, " ")[0])
		}
	}
	stopCommand1 := "kill"
	stopCommand2 := "-SIGTERM"
	cmd = exec.Command(stopCommand1, stopCommand2, processId)
	stderr = cmd.Run()
	if stderr != nil {
		emojlog.PrintLogMessage("kill was not successful (this is okay, vm supervisor will gracefully deal with it): "+stderr.Error(), emojlog.Warning)
	}

	emojlog.PrintLogMessage("Done stopping the VM: "+vmName, emojlog.Changed)
}

func vmSupervisorCleanup(vmName string) {
	emojlog.PrintLogMessage("Starting vm supervisor cleanup", emojlog.Debug)
	reMatchVm, _ := regexp.Compile(`for\s` + vmName + `\s&`)
	processId := ""

	iteration := 0
	for {
		time.Sleep(time.Second * 2)
		processId = ""
		cmd := exec.Command("pgrep", "-lf", vmName)
		stdout, stderr := cmd.Output()
		if stderr != nil {
			if cmd.ProcessState.ExitCode() == 1 {
				_ = 0
			} else {
				emojlog.PrintLogMessage("pgrep exited with an error: "+stderr.Error(), emojlog.Error)
			}
		}

		for _, v := range strings.Split(string(stdout), "\n") {
			v = strings.TrimSpace(v)
			if reMatchVm.MatchString(v) {
				processId = strings.Split(v, " ")[0]
			}
		}

		if len(processId) < 1 {
			emojlog.PrintLogMessage("VM process is already dead", emojlog.Debug)
			break
		}

		iteration = iteration + 1
		if iteration > 3 {
			stopCommand1 := "kill"
			stopCommand2 := "-SIGKILL"
			cmd := exec.Command(stopCommand1, stopCommand2, processId)
			stderr := cmd.Run()
			if stderr != nil {
				emojlog.PrintLogMessage("kill was not successful: "+stderr.Error(), emojlog.Error)

			}
			emojlog.PrintLogMessage("Forcefully killing the vm_supervisor, due to operation timeout: "+processId, emojlog.Debug)
		}
	}
	emojlog.PrintLogMessage("Done cleaning up after vm supervisor", emojlog.Changed)
}

func NetworkCleanup(vmName string, quiet bool) {
	if !quiet {
		emojlog.PrintLogMessage("Starting network cleanup", emojlog.Debug)
	}
	cmd := exec.Command("ifconfig")
	stdout, stderr := cmd.Output()
	if stderr != nil && !quiet {
		emojlog.PrintLogMessage("ifconfig exited with an error: "+stderr.Error(), emojlog.Error)
	}

	reMatchDescription, _ := regexp.Compile(`.*description:.*`)
	reMatchVm, _ := regexp.Compile(`\s+` + vmName + `\s+`)
	rePickTap, _ := regexp.Compile(`[\s|"]tap\d+`)
	for _, v := range strings.Split(string(stdout), "\n") {
		if reMatchDescription.MatchString(v) && reMatchVm.MatchString(v) {
			tap := rePickTap.FindString(v)
			tap = strings.TrimSpace(tap)
			tap = strings.ReplaceAll(tap, "\"", "")
			emojlog.PrintLogMessage("Destroying TAP interface: "+tap, emojlog.Debug)
			ifconfigDestroyCmd1 := "ifconfig"
			ifconfigDestroyCmd3 := "destroy"
			cmd := exec.Command(ifconfigDestroyCmd1, tap, ifconfigDestroyCmd3)
			stderr := cmd.Run()
			if stderr != nil && !quiet {
				emojlog.PrintLogMessage("ifconfig destroy was not successful: "+stderr.Error(), emojlog.Error)

			}
		}
	}
	if !quiet {
		emojlog.PrintLogMessage("Done cleaning up TAP network interfaces", emojlog.Debug)
	}
}

func BhyvectlDestroy(vmName string, quiet bool) {
	if !quiet {
		emojlog.PrintLogMessage("Cleaning up Bhyve resources", emojlog.Debug)
	}
	time.Sleep(time.Second * 2)
	lsCommand1 := "ls"
	lsCommand2 := "-1"
	lsCommand3 := "/dev/vmm/"
	cmd := exec.Command(lsCommand1, lsCommand2, lsCommand3)
	stdout, _ := cmd.Output()

	matchVM, _ := regexp.Compile(`^` + vmName + `$`)
	for _, v := range strings.Split(string(stdout), "\n") {
		v = strings.TrimSpace(v)
		if matchVM.MatchString(v) {
			if !quiet {
				emojlog.PrintLogMessage("Destroying a VM using bhyvectl: "+vmName, emojlog.Debug)
			}
			bhyvectlCommand1 := "bhyvectl"
			bhyvectlCommand2 := "--destroy"
			bhyvectlCommand3 := "--vm=" + vmName
			cmd := exec.Command(bhyvectlCommand1, bhyvectlCommand2, bhyvectlCommand3)
			stderr := cmd.Run()
			if stderr != nil && !quiet {
				emojlog.PrintLogMessage("bhyvectl exited with an error: "+stderr.Error(), emojlog.Error)
			}
		}
	}
	if !quiet {
		emojlog.PrintLogMessage("Done cleaning up Bhyve resources", emojlog.Changed)
	}
}
