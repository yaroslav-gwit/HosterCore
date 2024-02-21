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
func VmNetworkCleanup(vmName string) (r []Iface, e error) {
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
	reFindTapNew := regexp.MustCompile(`iface::.*?\s+`)
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

// This function creates a new TAP interface, sets the correct description for it,
// and returns it's (TAP interface) name back to the caller.
func CreateTapInterface(vmName string, networkName string) (r string, e error) {
	// Check if the network exists
	networks, err := GetNetworkConfig()
	if err != nil {
		e = err
		return
	}
	networkFound := false
	for _, v := range networks {
		if v.NetworkName == networkName {
			networkFound = true
		}
	}
	if !networkFound {
		e = fmt.Errorf("network with the name %s does not exist", networkName)
		return
	}
	// EOF Check if the network exists

	// Create new epair interface
	out, err := exec.Command("ifconfig", "tap", "create").CombinedOutput()
	if err != nil {
		e = fmt.Errorf("%s; %s", strings.TrimSpace(string(out)), err.Error())
		return
	}
	// EOF Create new epair interface

	// Set newly created interface name as a return value
	r = strings.TrimSpace(string(out))
	// EOF Set newly created interface name as a return value

	// Set a description for the new interface
	out, err = exec.Command("ifconfig", r, "description", fmt.Sprintf("\"vm::%s iface::%s network::%s\"", vmName, r, networkName)).CombinedOutput()
	if err != nil {
		e = fmt.Errorf("%s; %s", strings.TrimSpace(string(out)), err.Error())
		return
	}
	// EOF Set a description for the new interface

	// Add the interface to the VM network bridge
	out, err = exec.Command("ifconfig", "vm-"+networkName, "addm", r, "up").CombinedOutput()
	if err != nil {
		e = fmt.Errorf("%s; %s", strings.TrimSpace(string(out)), err.Error())
		return
	}
	// EOF Add the interface to the VM network bridge

	// Bring up the interface
	out, err = exec.Command("ifconfig", r, "up").CombinedOutput()
	if err != nil {
		e = fmt.Errorf("%s; %s", strings.TrimSpace(string(out)), err.Error())
		return
	}
	// EOF Bring up the interface

	return
}
