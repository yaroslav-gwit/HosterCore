package osfreebsd

import (
	"errors"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

// Returns a slice of strings from `dmesg.boot` split at a carriage return
func DmesgCpuGrep() ([]string, error) {
	out, err := exec.Command("grep", "-i", "package", "/var/run/dmesg.boot").CombinedOutput()
	if err != nil {
		errorString := strings.TrimSpace(string(out)) + "; " + err.Error()
		return []string{}, errors.New(errorString)
	}

	r := []string{}
	for _, v := range strings.Split(string(out), "\n") {
		v = strings.TrimSpace(v)
		if len(v) > 0 {
			r = append(r, v)
		}
	}

	return r, nil
}

type PgrepPID struct {
	ProcessId  int
	ProcessCmd string
}

// Searches for a given process on your system, using a command start string
//
// Returns a struct with process ID as int and the command used to start it (string):
func Pgrep(processName string) (pids []PgrepPID, finalErr error) {
	// Clean the input
	processName = strings.TrimSpace(processName)
	reMatchInputFilter := regexp.MustCompile(`'|"`)
	processName = reMatchInputFilter.ReplaceAllString(processName, "")

	out, err := exec.Command("/bin/pgrep", "-afSl", processName).CombinedOutput()
	if err != nil {
		errorString := strings.TrimSpace(string(out)) + "; " + err.Error()
		finalErr = errors.New(errorString)
		return
	}

	reSplitSpace := regexp.MustCompile(`\s+`)
	for _, v := range strings.Split(string(out), "\n") {
		v = strings.TrimSpace(v)
		if len(v) < 1 {
			continue
		}

		pidSplit := reSplitSpace.Split(v, -1)
		pidNum, err := strconv.Atoi(pidSplit[0])

		if err != nil {
			finalErr = err
			return
		}

		processCmd := ""
		if len(pidSplit) >= 2 {
			processCmd = strings.Join(pidSplit[1:len(pidSplit)-1], " ")
			processCmd = strings.TrimSpace(processCmd)
		} else if len(pidSplit) >= 1 {
			processCmd = pidSplit[1]
			processCmd = strings.TrimSpace(processCmd)
		} else {
			finalErr = errors.New("could not find a process cmd string in " + v)
			return
		}

		pids = append(pids, PgrepPID{ProcessId: pidNum, ProcessCmd: processCmd})
	}

	return
}
