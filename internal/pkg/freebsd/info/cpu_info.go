// Copyright 2023 Hoster Authors. All rights reserved.
// Use of this source code is governed by an Apache License 2.0
// license that can be found in the LICENSE file.

package FreeBSDOsInfo

import (
	FreeBSDsysctls "HosterCore/internal/pkg/freebsd/sysctls"
	"errors"
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

type CpuInfo struct {
	Model        string `json:"cpu_model"`
	Architecture string `json:"cpu_arch"`
	Sockets      int    `json:"cpu_sockets"`
	Cores        int    `json:"cpu_cores"`
	Threads      int    `json:"cpu_threads"`
	OverallCpus  int    `json:"overall_cpus"`
}

// Returns a structured CPU information for your FreeBSD system
func GetCpuInfo() (r CpuInfo, e error) {
	reMatchNumber := regexp.MustCompile(`\d+`)
	dmesg, err := DmesgCpuGrep()
	if err != nil {
		e = err
		return
	}

	reMatchSockets := regexp.MustCompile(`\s+(\d+)\s+package`)
	socketsString := reMatchSockets.FindString(dmesg[0])
	socketsString = reMatchNumber.FindString(socketsString)
	r.Sockets, err = strconv.Atoi(socketsString)
	if err != nil {
		e = fmt.Errorf("socket err: %s", err.Error())
		return
	}

	reMatchThreads := regexp.MustCompile(`x\s+(\d+)\s+(?:hardware\s+)?threads`)
	threadsString := reMatchThreads.FindString(dmesg[0])
	threadsString = reMatchNumber.FindString(threadsString)
	r.Threads, err = strconv.Atoi(threadsString)
	if err != nil {
		r.Threads = 1
	}

	allCpus, err := FreeBSDsysctls.SysctlHwNcpu()
	if err != nil {
		e = err
		return
	}
	r.Cores = allCpus / (r.Threads * r.Sockets)

	r.Model, err = FreeBSDsysctls.SysctlHwModel()
	if err != nil {
		e = err
		return
	}

	r.Architecture, err = FreeBSDsysctls.SysctlHwMachine()
	if err != nil {
		return r, err
	}

	r.OverallCpus, err = FreeBSDsysctls.SysctlHwNcpu()
	if err != nil {
		e = err
		return
	}

	return
}

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
