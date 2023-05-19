package main

import (
	"bufio"
	"hoster/cmd"
	"io"
	"log"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"
	"sync"
)

func main() {
	// Get env vars passed from "hoster vm start"
	vmStartCommand := os.Getenv("VM_START")
	logFileLocation := os.Getenv("LOG_FILE")
	vmName := os.Getenv("VM_NAME")

	// Create or open the log file for writing
	logFile, err := os.OpenFile(logFileLocation, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0640)
	if err != nil {
		log.Fatal("Unable to open log file: " + err.Error())
	}
	// Redirect the output of log.Fatal to the log file
	log.SetOutput(logFile)

	// Start the process
	parts := strings.Fields(vmStartCommand)
	for {
		log.Println("[stdout] Starting the VM as a child process")
		hupCmd := exec.Command(parts[0], parts[1:]...)
		stdout, err := hupCmd.StdoutPipe()
		if err != nil {
			log.Fatalf("[stderr] Failed to create stdout pipe: %v", err)
		}
		stderr, err := hupCmd.StderrPipe()
		if err != nil {
			log.Fatalf("[stderr] Failed to create stderr pipe: %v", err)
		}

		var wg sync.WaitGroup
		wg.Add(2)

		stdoutReader := bufio.NewReader(stdout)
		go func() {
			defer wg.Done()
			readAndLogOutput(stdoutReader, "stdout")
		}()

		stderrReader := bufio.NewReader(stderr)
		go func() {
			defer wg.Done()
			readAndLogOutput(stderrReader, "stderr")
		}()

		done := make(chan error)
		startVmProcess(hupCmd, done)
		wg.Wait()

		if err := <-done; err != nil {
			log.Printf("[stdout] VM child process ended: " + err.Error())
			if exitError, ok := err.(*exec.ExitError); ok {
				if status, ok := exitError.Sys().(interface{ ExitStatus() int }); ok {
					exitCode := status.ExitStatus()
					// 
					// List of Bhyve's exit codes (not confirmed yet!):
					// 0: The virtual machine terminated normally without any errors.
					// 1: The virtual machine terminated with an unspecified error.
					// 2: The virtual machine terminated due to a fatal error or panic.
					// 3: The virtual machine was reset or rebooted.
					// 4: The virtual machine was terminated due to a guest-initiated shutdown or ACPI power button press.
					// 5: The virtual machine was terminated due to a host-initiated shutdown or request.
					// 6: The virtual machine was terminated due to a watchdog timeout.
					// 7: The virtual machine was terminated due to an internal Bhyve error.
					// 8: The virtual machine was terminated due to an invalid command-line argument or configuration.
					// 9: The virtual machine was terminated due to an external signal (e.g., SIGINT, SIGTERM).
					// 
					if exitCode == 3 {
						// 
						// The below should fix a bug where VM goes unresponsive after 3-5 reboots
						// executed from within the VM itself
						// 
						// The idea is to kill current VM_SUPERVISOR parent and it's child process,
						// then start a new one
						// 
						// The fix would also allow to pick up config changes across the
						// regular VM reboots
						// 
						log.Println("[stdout] Bhyve received a reboot signal: " + strconv.Itoa(exitCode) + ". Rebooting...")
						cmd.NetworkCleanup(vmName, true)
						cmd.BhyvectlDestroy(vmName, true)
						restartVmProcess(vmName)
						os.Exit(0)
					} else if exitCode == 4 || exitCode == 5 {
						log.Println("[stdout] Bhyve received a shutdown signal: " + strconv.Itoa(exitCode) + ". Shutting down...")
						log.Println("[stdout] Performing network cleanup")
						cmd.NetworkCleanup(vmName, true)
						log.Println("[stdout] Performing Bhyve cleanup")
						cmd.BhyvectlDestroy(vmName, true)
						os.Exit(0)
					} else {
						log.Println("[stderr] Bhyve returned a panic exit code: " + strconv.Itoa(exitCode))
						log.Println("[stdout] Shutting down all VM related processes and performing system clean up")
						cmd.NetworkCleanup(vmName, true)
						cmd.BhyvectlDestroy(vmName, true)
						os.Exit(101)
					}
				}
			}
			os.Exit(0)
		}
	}
}

func readAndLogOutput(reader *bufio.Reader, name string) {
	for {
		line, err := reader.ReadString('\n')
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("Failed to read %s: %v", name, err)
		}
		line = strings.TrimSpace(line)
		if line != "" {
			log.Printf("[%s] %s\n", name, line)
		}
	}
}

func startVmProcess(cmd *exec.Cmd, done chan error) {
	if err := cmd.Start(); err != nil {
		log.Fatalf("Failed to start command: %v", err)
	}
	go func() {
		done <- cmd.Wait()
	}()
}

func restartVmProcess(vmName string) {
	execPath, err := os.Executable()
	if err != nil {
		log.Fatalf("Failed to start command: %v", err)
	}
	execFile := path.Dir(execPath) + "/hoster"
	out, err := exec.Command("nohup", execFile, "start", vmName).CombinedOutput()
	if err != nil {
		log.Fatal("Could not restart the VM: " + string(out) + "; " + err.Error())
	}
}
