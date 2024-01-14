package FreeBSDPgrep

import (
	"errors"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

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
		finalErr = errors.New("Pgrep() " + errorString)
		return
	}

	reSplitSpace := regexp.MustCompile(`\s+`)
	for _, v := range strings.Split(string(out), "\n") {
		v = strings.TrimSpace(v)
		if len(v) < 1 {
			continue
		}

		pidSplit := reSplitSpace.Split(v, 2)
		pidNum, err := strconv.Atoi(pidSplit[0])
		if err != nil {
			finalErr = errors.New("Pgrep() " + err.Error())
			return
		}

		processCmd := ""
		if len(pidSplit) > 0 {
			processCmd = pidSplit[1]
			processCmd = strings.TrimSpace(processCmd)
		} else {
			finalErr = errors.New("Pgrep() could not find a process cmd string in " + v)
			return
		}

		pids = append(pids, PgrepPID{ProcessId: pidNum, ProcessCmd: processCmd})
	}

	return
}
