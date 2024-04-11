// Copyright 2024 Hoster Authors. All rights reserved.
// Use of this source code is governed by an Apache License 2.0
// license that can be found in the LICENSE file.

package main

import (
	HosterNetwork "HosterCore/internal/pkg/hoster/network"
	HosterVm "HosterCore/internal/pkg/hoster/vm"
	HosterVmUtils "HosterCore/internal/pkg/hoster/vm/utils"
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"sync"

	"github.com/sirupsen/logrus"
)

var vmName string
var version = "" // automatically set during the build process

func main() {
	// Print the version and exit
	args := os.Args
	if len(args) > 1 {
		res := os.Args[1]
		if res == "version" || res == "v" || res == "--version" || res == "-v" {
			fmt.Println(version)
			return
		}
	}

	// Get env vars passed from "hoster vm start"
	vmName = os.Getenv("VM_NAME")
	vmStartCommand := os.Getenv("VM_START")

	// Start the process
	parts := strings.Fields(vmStartCommand)
	for {
		log.WithFields(logrus.Fields{"type": LOG_SUPERVISOR}).Info("SUPERVISED SESSION STARTED: VM boot process has been initiated")
		hupCmd := exec.Command(parts[0], parts[1:]...)
		stdout, err := hupCmd.StdoutPipe()
		if err != nil {
			log.WithFields(logrus.Fields{"type": LOG_SUPERVISOR}).Error("Failed to create stdout pipe: " + err.Error())
			os.Exit(101)
		}
		stderr, err := hupCmd.StderrPipe()
		if err != nil {
			log.WithFields(logrus.Fields{"type": LOG_SUPERVISOR}).Error("Failed to create stderr pipe: " + err.Error())
			os.Exit(102)
		}

		var wg sync.WaitGroup
		wg.Add(2)

		stdoutReader := bufio.NewReader(stdout)
		go func() {
			defer wg.Done()
			readAndLogOutput(stdoutReader, LOG_SYS_OUT)
		}()

		stderrReader := bufio.NewReader(stderr)
		go func() {
			defer wg.Done()
			readAndLogOutput(stderrReader, LOG_SYS_ERR)
		}()

		done := make(chan error)
		startVmProcess(hupCmd, done)
		wg.Wait()

		processErr := <-done
		if processErr != nil {
			log.WithFields(logrus.Fields{"type": LOG_SUPERVISOR}).Error("VM child process ended with a non-zero exit code: " + processErr.Error())
		}

		processExitStatus, correctReturnType := processErr.(*exec.ExitError)
		if correctReturnType {
			exitCode := processExitStatus.ProcessState.ExitCode()
			if exitCode == 1 || exitCode == 2 {
				log.WithFields(logrus.Fields{"type": LOG_SUPERVISOR}).Infof("Bhyve received a shutdown signal: %d. Executing the shutdown sequence...", exitCode)

				log.WithFields(logrus.Fields{"type": LOG_SUPERVISOR}).Info("Shutting down -> Performing network cleanup")
				_, _ = HosterNetwork.VmNetworkCleanup(vmName)
				log.WithFields(logrus.Fields{"type": LOG_SUPERVISOR}).Info("Shutting down -> Performing Bhyve cleanup")
				_ = HosterVmUtils.BhyveCtlDestroy(vmName)

				log.WithFields(logrus.Fields{"type": LOG_SUPERVISOR}).Info("SUPERVISED SESSION ENDED. The VM has been shutdown.")
				os.Exit(0)
			} else {
				log.WithFields(logrus.Fields{"type": LOG_SUPERVISOR}).Errorf("Bhyve returned a panic exit code: %d. Shutting down all VM related processes and performing system clean up.", exitCode)
				_, _ = HosterNetwork.VmNetworkCleanup(vmName)
				_ = HosterVmUtils.BhyveCtlDestroy(vmName)
				log.WithFields(logrus.Fields{"type": LOG_SUPERVISOR}).Error("SUPERVISED SESSION ENDED. Unexpected exit code.")
				os.Exit(101)
			}
		} else {
			log.WithFields(logrus.Fields{"type": LOG_SUPERVISOR}).Info("Bhyve received a reboot signal. Executing the reboot sequence...")

			log.WithFields(logrus.Fields{"type": LOG_SUPERVISOR}).Info("Rebooting -> Performing network cleanup")
			_, _ = HosterNetwork.VmNetworkCleanup(vmName)
			log.WithFields(logrus.Fields{"type": LOG_SUPERVISOR}).Info("Rebooting -> Performing Bhyve cleanup")
			_ = HosterVmUtils.BhyveCtlDestroy(vmName)

			log.WithFields(logrus.Fields{"type": LOG_SUPERVISOR}).Info("Rebooting -> Performing Bhyve cleanup")
			restartVmProcess(vmName)
			os.Exit(0)
		}

		log.WithFields(logrus.Fields{"type": LOG_SUPERVISOR}).Error("SUPERVISED SESSION ENDED. SOMETHING UNPREDICTED HAPPENED! THE PROCESS HAD TO EXIT!")
		_, _ = HosterNetwork.VmNetworkCleanup(vmName)
		_ = HosterVmUtils.BhyveCtlDestroy(vmName)
		os.Exit(1000)
	}
}

func readAndLogOutput(reader *bufio.Reader, name string) {
	reMatchProcessFailureLogLine1 := regexp.MustCompile(`read\s+\|0:\s+file\s+already\s+closed`) // match this result: "read |0: file already closed"
	for {
		line, err := reader.ReadString('\n')
		if err == io.EOF {
			break
		}
		if err != nil {
			log.WithFields(logrus.Fields{"type": LOG_SUPERVISOR}).Error(name + "; " + err.Error())
			os.Exit(100)
		}

		line = strings.TrimSpace(line)
		if line != "" {
			log.WithFields(logrus.Fields{"type": name}).Info(line)
		}

		// Match one of the error outputs, and kill the VM process if something has gone wrong. Aka "fail early" principle.
		if reMatchProcessFailureLogLine1.MatchString(line) {
			_, _ = HosterNetwork.VmNetworkCleanup(vmName)
			_ = HosterVmUtils.BhyveCtlDestroy(vmName)
			log.WithFields(logrus.Fields{"type": LOG_SUPERVISOR}).Error("SUPERVISED SESSION ENDED. Bhyve process failure.")
			os.Exit(1001)
		}
	}
}

func startVmProcess(cmd *exec.Cmd, done chan error) {
	err := cmd.Start()
	if err != nil {
		log.WithFields(logrus.Fields{"type": LOG_SUPERVISOR}).Error("Failed to start the VM using bhyve: " + err.Error())
		os.Exit(100)
	}
	go func() {
		done <- cmd.Wait()
	}()
}

func restartVmProcess(vmName string) {
	err := HosterVm.Start(vmName, false, false)
	if err != nil {
		log.WithFields(logrus.Fields{"type": LOG_SUPERVISOR}).Error("Failed to restart the VM: " + err.Error())
		os.Exit(101)
	}
}
