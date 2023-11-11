package cmd

import (
	"HosterCore/emojlog"
	"errors"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"golang.org/x/exp/slices"
)

var (
	vmStopCmdForceStop bool
	vmStopCmdCleanUp   bool
	vmStopCmd          = &cobra.Command{
		Use:   "stop [vmName]",
		Short: "Stop a particular VM using it's name",
		Long:  `Stop a particular VM using it's name`,
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			err := checkInitFile()
			if err != nil {
				emojlog.PrintLogMessage(err.Error(), emojlog.Error)
			}

			err = VmStop(args[0], vmStopCmdForceStop, vmStopCmdCleanUp)
			if err != nil {
				emojlog.PrintLogMessage(err.Error(), emojlog.Error)
			}
		},
	}
)

func VmStop(vmName string, forceKill bool, forceCleanup bool) error {
	allVms := getAllVms()
	if !slices.Contains(allVms, vmName) {
		return errors.New("VM is not found on this system")
	} else if !VmLiveCheck(vmName) && !forceKill {
		return errors.New("VM is already stopped")
	}

	err := SendShutdownSignalToVm(vmName, forceKill, true, forceCleanup)
	if err != nil {
		return err
	}

	// OLD WAY OF STOPPING VMs
	// StopBhyveProcess(vmName, false, forceStop)
	// vmSupervisorCleanup(vmName)
	// StopBhyveProcess(vmName, true, forceKill)
	// NetworkCleanup(vmName, false)
	// BhyvectlDestroy(vmName, false)

	return nil
}

func SendShutdownSignalToVm(vmName string, forceKill bool, logOutput bool, cleanup bool) error {
	if logOutput {
		emojlog.PrintLogMessage("Stopping the VM: "+vmName, emojlog.Debug)
	}

	pid, err := exec.Command("pgrep", "-afSl", vmName).CombinedOutput()
	if err != nil && !forceKill {
		return errors.New("could not find the VM process specified (is the VM running?)")
	}

	reMatchVm := regexp.MustCompile(`\s+bhyve:\s+` + vmName + `$`)
	reSplitSpace := regexp.MustCompile(`\s+`)
	vmPid := ""
	for _, v := range strings.Split(string(pid), "\n") {
		if reMatchVm.MatchString(v) {
			vmPid = reSplitSpace.Split(v, -1)[0]
			vmPid = strings.TrimSpace(vmPid)
		}
	}

	if len(vmPid) == 0 && !forceKill {
		return errors.New("could not find the VM process specified (is the VM running?)")
	}

	if forceKill {
		err = exec.Command("kill", "-SIGKILL", vmPid).Run()
		if logOutput && err == nil && len(vmPid) > 0 {
			emojlog.PrintLogMessage("Forceful SIGKILL signal has been sent to: "+vmName+"; PID: "+vmPid, emojlog.Changed)
		}

		// Clean-up some leftover artifacts
		if cleanup {
			_ = exec.Command("bhyvectl", "--destroy", "--vm="+vmName).Run()
			vmSupervisorCleanup(vmName, false)
			NetworkCleanup(vmName, false)
			BhyvectlDestroy(vmName, false)
		}
	} else {
		_ = exec.Command("kill", "-SIGTERM", vmPid).Run()
		if logOutput {
			emojlog.PrintLogMessage("Graceful SIGTERM signal has been sent to: "+vmName+"; PID: "+vmPid, emojlog.Changed)
		}
	}

	return nil
}

func StopBhyveProcess(vmName string, quiet bool, kill bool) {
	if !quiet {
		emojlog.PrintLogMessage("Stopping the VM: "+vmName, emojlog.Info)
	}

	cmd := exec.Command("pgrep", "-lf", vmName)
	stdout, stderr := cmd.Output()
	if stderr != nil {
		if cmd.ProcessState.ExitCode() == 1 {
			_ = 0
		} else if !quiet {
			emojlog.PrintLogMessage("pgrep exited with an error: "+stderr.Error(), emojlog.Error)
		}
	}

	processId := ""
	reMatchVm, _ := regexp.Compile(`.*bhyve:\s` + vmName)
	reMatchVmX2, _ := regexp.Compile(`(.*` + vmName + `\s){2}`)

	for _, v := range strings.Split(string(stdout), "\n") {
		if reMatchVm.MatchString(v) {
			processId = strings.TrimSpace(strings.Split(v, " ")[0])
			break
		} else if reMatchVmX2.MatchString(v) {
			processId = strings.TrimSpace(strings.Split(v, " ")[0])
			break
		}
	}

	stopCommand1 := ""
	stopCommand2 := ""
	if kill {
		stopCommand1 = "kill"
		stopCommand2 = "-SIGKILL"
	} else {
		stopCommand1 = "kill"
		stopCommand2 = "-SIGTERM"
	}

	cmd = exec.Command(stopCommand1, stopCommand2, processId)
	stderr = cmd.Run()
	if stderr != nil && !quiet {
		emojlog.PrintLogMessage("kill was not successful (this is okay, vm supervisor will gracefully deal with it): "+stderr.Error(), emojlog.Warning)
	}

	if !quiet {
		emojlog.PrintLogMessage("Done stopping the bhyve VM process: "+vmName, emojlog.Changed)
	}
}

func vmSupervisorCleanup(vmName string, logOutput bool) {
	if logOutput {
		emojlog.PrintLogMessage("Performing vm_supervisor cleanup", emojlog.Debug)
	}
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
			} else if logOutput {
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
			if logOutput {
				emojlog.PrintLogMessage("VM process is dead", emojlog.Debug)
			}
			break
		}

		iteration = iteration + 1
		if iteration > 3 {
			cmd := exec.Command("kill", "-SIGKILL", processId)
			stderr := cmd.Run()

			if logOutput {
				if stderr != nil {
					emojlog.PrintLogMessage("kill was not successful: "+stderr.Error(), emojlog.Error)
				}
				emojlog.PrintLogMessage("Forcefully killing the vm_supervisor, due to operation timeout: "+processId, emojlog.Debug)
			}
		}
	}
	if logOutput {
		emojlog.PrintLogMessage("Done cleaning up after vm supervisor", emojlog.Changed)
	}
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
