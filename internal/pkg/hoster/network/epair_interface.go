// Copyright 2023 Hoster Authors. All rights reserved.
// Use of this source code is governed by an Apache License 2.0
// license that can be found in the LICENSE file.

package HosterNetwork

import (
	"fmt"
	"os/exec"
	"strings"
)

type EpairInterface struct {
	IFaceA string
	IFaceB string
}

func CreateEpairInterface(jailName string, networkName string) (r EpairInterface, e error) {
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
	out, err := exec.Command("ifconfig", "epair", "create").CombinedOutput()
	if err != nil {
		e = fmt.Errorf("%s; %s", strings.TrimSpace(string(out)), err.Error())
		return
	}
	// EOF Create new epair interface

	// Add newly created interfaces to the return list
	r.IFaceA = strings.TrimSpace(string(out))
	r.IFaceB = strings.TrimSuffix(r.IFaceA, "a") + "b"
	// EOF Add newly created interfaces to the return list

	// Set a description for the new interface
	out, err = exec.Command("ifconfig", r.IFaceA, "description", fmt.Sprintf("\"%s %s network:%s\"", jailName, r.IFaceA, networkName)).CombinedOutput()
	if err != nil {
		e = fmt.Errorf("%s; %s", strings.TrimSpace(string(out)), err.Error())
		return
	}
	// EOF Set a description for the new interface

	// Add the interface to the VM network bridge
	out, err = exec.Command("ifconfig", "vm-"+networkName, "addm", r.IFaceA, "up").CombinedOutput()
	if err != nil {
		e = fmt.Errorf("%s; %s", strings.TrimSpace(string(out)), err.Error())
		return
	}
	// EOF Add the interface to the VM network bridge

	// Bring up the interface
	out, err = exec.Command("ifconfig", r.IFaceA, "up").CombinedOutput()
	if err != nil {
		e = fmt.Errorf("%s; %s", strings.TrimSpace(string(out)), err.Error())
		return
	}
	// EOF Bring up the interface

	return
}
