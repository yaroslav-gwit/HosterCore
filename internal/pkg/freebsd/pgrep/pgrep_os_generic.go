// Copyright 2024 Hoster Authors. All rights reserved.
// Use of this source code is governed by an Apache License 2.0
// license that can be found in the LICENSE file.

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

// Searches for a given process on your system, using the start command string (e.g. "bash script.sh", or "/opt/custom/binary --bla").
//
// Returns a list of structs with the process ID (int) and the command (string) used to start it.
//
// TIP: use a generic string to begin with (e.g. "bhyve") and then filter the results using regex or string matching patterns inside of your caller function.
func Pgrep(processName string) (r []PgrepPID, e error) {
	// Clean the input
	processName = strings.TrimSpace(processName)
	reMatchInputFilter := regexp.MustCompile(`'|"`)
	processName = reMatchInputFilter.ReplaceAllString(processName, "")

	out, err := exec.Command("/bin/pgrep", "-afSl", processName).CombinedOutput()
	if err != nil {
		errorString := strings.TrimSpace(string(out)) + "; " + err.Error()
		e = errors.New("Pgrep() " + errorString)
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
			e = errors.New("Pgrep() " + err.Error())
			return
		}

		processCmd := ""
		if len(pidSplit) > 0 {
			processCmd = pidSplit[1]
			processCmd = strings.TrimSpace(processCmd)
		} else {
			e = errors.New("Pgrep() could not find a process cmd string in " + v)
			return
		}

		r = append(r, PgrepPID{ProcessId: pidNum, ProcessCmd: processCmd})
	}

	return
}
