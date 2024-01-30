// Copyright 2023 Hoster Authors. All rights reserved.
// Use of this source code is governed by an Apache License 2.0
// license that can be found in the LICENSE file.

package HosterNetwork

import (
	"fmt"
	"os/exec"
	"regexp"
	"strings"
)

// TAP Interface Struct used in the VM network clean-up (vm stop) procedure.
type Iface struct {
	IfaceName string // TAP Interface Name
	Success   bool   // TAP Interface removal was successful
	Failure   bool   // TAP Interface removal has failed
}

// Perform a network clean-up, and return a list of networks interfaces.
// The return value is a struct that includes the interface name, and of the destroy op was a success.
func NetworkCleanup(vmName string) (r []Iface, e error) {
	out, err := exec.Command("ifconfig").CombinedOutput()
	if err != nil {
		e = fmt.Errorf("error: %s; %s", strings.TrimSpace(string(out)), err.Error())
		return
	}

	outSplit := strings.Split(string(out), "\n")
	reMatchDescription, _ := regexp.Compile(`description:\s+`)

	// Old format network description
	reMatchVm := regexp.MustCompile(`\s+` + vmName + `\s+`)
	reFindTap := regexp.MustCompile(`\"tap\d+`)
	for _, v := range outSplit {
		v = strings.TrimSpace(v)
		if reMatchDescription.MatchString(v) && reMatchVm.MatchString(v) {
			v = reFindTap.FindString(v)
			v = strings.TrimPrefix(v, `"`)
			v = strings.TrimSpace(v)
			r = append(r, Iface{IfaceName: v})
		}
	}

	// New format network description
	reMatchVmNew := regexp.MustCompile(`vm::` + vmName)
	reFindTapNew := regexp.MustCompile(`iface::.*?\s+` + vmName)
	for _, v := range outSplit {
		v = strings.TrimSpace(v)
		if reMatchDescription.MatchString(v) && reMatchVmNew.MatchString(v) {
			v = reFindTapNew.FindString(v)
			v = strings.TrimPrefix(v, `iface::`)
			v = strings.TrimSpace(v)
			r = append(r, Iface{IfaceName: v})
		}
	}

	// Loop over the list of the interfaces that match our description, and destroy them.
	for i, v := range r {
		err := exec.Command("ifconfig", v.IfaceName, "destroy").Run()
		if err != nil {
			r[i].Failure = true
		} else {
			r[i].Success = true
		}
	}

	return
}
