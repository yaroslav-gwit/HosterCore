package main

import (
	"bufio"
	"hoster/cmd"
	"io"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"
	"sync"
	"time"
)

var logFileLocation string
func main() {
	// Get env vars passed from "hoster vm start"
	vmStartCommand := os.Getenv("VM_START")
	vmName := os.Getenv("VM_NAME")
	logFileLocation = os.Getenv("LOG_FILE")

	// Create or open the log file for writing
	// logFile, err := os.OpenFile(logFileLocation, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0640)
	// if err != nil {
	// 	log.Fatal("Unable to open log file: " + err.Error())
	// }
	// log.SetOutput(logFile)

	// Start the process
	parts := strings.Fields(vmStartCommand)
	for {
		logFileOutput("stdout", "Starting the VM as a child process")
		hupCmd := exec.Command(parts[0], parts[1:]...)
		stdout, err := hupCmd.StdoutPipe()
		if err != nil {
			logFileOutput("stderr", "Failed to create stdout pipe: " + err.Error())
			os.Exit(101)
		}
		stderr, err := hupCmd.StderrPipe()
		if err != nil {
			logFileOutput("stderr", "Failed to create stderr pipe: " + err.Error())
			os.Exit(102)
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

		processErr := <-done
		if processErr != nil {
			logFileOutput("stdout", "VM child process ended with a non-zero exit code: " + processErr.Error())
		}
		if exitError, ok := processErr.(*exec.ExitError); ok {
			exitCode := exitError.ProcessState.ExitCode()
			if exitCode == 1 || exitCode == 2 {
				logFileOutput("stdout", "Bhyve received a shutdown signal: " + strconv.Itoa(exitCode) + ". Shutting down...")
				logFileOutput("stdout", "Performing network cleanup")
				cmd.NetworkCleanup(vmName, true)
				logFileOutput("stdout", "Performing Bhyve cleanup")
				cmd.BhyvectlDestroy(vmName, true)
				logFileOutput("notice", "ALL CLEANUP PROCEDURES ARE DONE.")
				os.Exit(0)
			} else {
				logFileOutput("stderr", "Bhyve returned a panic exit code: " + strconv.Itoa(exitCode))
				logFileOutput("stderr", "Shutting down all VM related processes and performing system clean up")
				cmd.NetworkCleanup(vmName, true)
				cmd.BhyvectlDestroy(vmName, true)
				logFileOutput("notice", "ALL CLEANUP PROCEDURES ARE DONE.")
				os.Exit(101)
			}
		} else {
			logFileOutput("stdout", "Bhyve received a reboot signal. Rebooting...")
			cmd.NetworkCleanup(vmName, true)
			cmd.BhyvectlDestroy(vmName, true)
			restartVmProcess(vmName)
			logFileOutput("notice", "ALL DONE. WILL START THE VM AGAIN SHORTLY.")
			os.Exit(0)
		}

		logFileOutput("notice", "SOMETHING UNPREDICTED HAPPENED! THE PROCESS HAD TO EXIT!")
		os.Exit(1000)
	}
}

func readAndLogOutput(reader *bufio.Reader, name string) {
	for {
		line, err := reader.ReadString('\n')
		if err == io.EOF {
			break
		}
		if err != nil {
			logFileOutput(name, err.Error())
			os.Exit(100)
		}
		line = strings.TrimSpace(line)
		if line != "" {
			logFileOutput(name, line)
		}
	}
}

func startVmProcess(cmd *exec.Cmd, done chan error) {
	err := cmd.Start()
	if err != nil {
		logFileOutput("stderr", "Failed to start command: "+ err.Error())
		os.Exit(100)
	}
	go func() {
		done <- cmd.Wait()
	}()
}

func restartVmProcess(vmName string) {
	execPath, err := os.Executable()
	if err != nil {
		logFileOutput("stderr", "Could not find the executable path: "+ err.Error())
		os.Exit(100)
	}
	execFile := path.Dir(execPath) + "/hoster"
	out, err := exec.Command("nohup", execFile, "start", vmName).CombinedOutput()
	if err != nil {
		logFileOutput("stderr", "Could not restart the VM: " + string(out) + "; " + err.Error())
		os.Exit(100)
	}
}

func logFileOutput(msgType string, msgString string) {
	// Create or open the log file for writing
	timeNow := time.Now().Format("2006-01-02 15:04:05")
	logFile, err := os.OpenFile(logFileLocation, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0640)
	if err != nil {
		_ = exec.Command("logger", err.Error()).Run()
	}
	// log.SetOutput(logFile)
	defer logFile.Close()

	// Append the line to the file
	_, err = logFile.WriteString(timeNow + " ["+msgType+"] " + msgString + "\n")
	if err != nil {
		_ = exec.Command("logger", err.Error()).Run()
	}
}
