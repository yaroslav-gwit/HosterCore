package main

import (
	"fmt"
	"os/exec"
	"strings"
)

func parseIfconfig() (r string, e error) {
	out, err := exec.Command("ifconfig").CombinedOutput()
	if err != nil {
		e = fmt.Errorf("could not run ifconfig: %s; %s", strings.TrimSpace(string(out)), err.Error())
		return
	}

	matchString := "carp:"
	for _, v := range strings.Split(string(out), "\n") {
		if strings.Contains(v, matchString) {
			r = strings.TrimSpace(v)
			return
		}
	}

	return
}
