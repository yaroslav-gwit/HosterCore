// Copyright 2023 Hoster Authors. All rights reserved.
// Use of this source code is governed by an Apache License 2.0
// license that can be found in the LICENSE file.

package FreeBSDps

import (
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

type ProcessTime struct {
	StartTime int64
	Command   string
}

func ProcessTimes() (r []ProcessTime, e error) {
	out, err := exec.Command("ps", "axwww", "-o", "etimes,command").CombinedOutput()
	if err != nil {
		e = fmt.Errorf("%s; %s", strings.TrimSpace(string(out)), err.Error())
		return
	}

	reSplitSpace := regexp.MustCompile(`\s+`)
	for _, v := range strings.Split(string(out), "\n") {
		v = strings.TrimSpace(v)

		split := reSplitSpace.Split(v, -1)
		if len(split) < 1 {
			continue
		}

		etime, err := strconv.ParseInt(split[0], 10, 64)
		if err != nil {
			continue
		}

		command := strings.TrimSpace(v)
		command = strings.TrimPrefix(command, fmt.Sprintf("%d", etime))
		command = strings.TrimSpace(command)
		r = append(r, ProcessTime{StartTime: etime, Command: command})
	}

	return
}
