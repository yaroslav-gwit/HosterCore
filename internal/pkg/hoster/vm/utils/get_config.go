// Copyright 2023 Hoster Authors. All rights reserved.
// Use of this source code is governed by an Apache License 2.0
// license that can be found in the LICENSE file.

package HosterVmUtils

import (
	FileExists "HosterCore/internal/pkg/file_exists"
	"encoding/json"
	"errors"
	"os"
	"strings"
)

type VmDisk struct {
	DiskType     string `json:"disk_type"`
	DiskLocation string `json:"disk_location"`
	DiskImage    string `json:"disk_image"`
	Comment      string `json:"comment"`
}

type VmNetwork struct {
	NetworkAdaptorType string `json:"network_adaptor_type"`
	NetworkBridge      string `json:"network_bridge"`
	NetworkMac         string `json:"network_mac"`
	IPAddress          string `json:"ip_address"`
	Comment            string `json:"comment"`
}

type VmSshKey struct {
	KeyOwner string `json:"key_owner"`
	KeyValue string `json:"key_value"`
	Comment  string `json:"comment"`
}

type Virtio9P struct {
	ShareName     string `json:"share_name"`
	ShareLocation string `json:"share_location"`
	ReadOnly      bool   `json:"read_only"`
}

type VmConfig struct {
	CPUSockets             string      `json:"cpu_sockets"`
	CPUCores               string      `json:"cpu_cores"`
	CPUThreads             int         `json:"cpu_threads,omitempty"`
	Memory                 string      `json:"memory"`
	Loader                 string      `json:"loader"`
	LiveStatus             string      `json:"live_status"`
	OsType                 string      `json:"os_type"`
	OsComment              string      `json:"os_comment"`
	Owner                  string      `json:"owner"`
	ParentHost             string      `json:"parent_host"`
	Networks               []VmNetwork `json:"networks"`
	Disks                  []VmDisk    `json:"disks"`
	IncludeHostwideSSHKeys bool        `json:"include_hostwide_ssh_keys"`
	VmSshKeys              []VmSshKey  `json:"vm_ssh_keys"`
	VncPort                string      `json:"vnc_port"`
	VncPassword            string      `json:"vnc_password"`
	Description            string      `json:"description"`
	UUID                   string      `json:"uuid,omitempty"`
	VGA                    string      `json:"vga,omitempty"`
	Passthru               []string    `json:"passthru,omitempty"`
	DisableXHCI            bool        `json:"disable_xhci,omitempty"`
	VncResolution          int         `json:"vnc_resolution,omitempty"`
	Shares                 []Virtio9P  `json:"9p_shares,omitempty"`
}

// Reads and returns the vm_config.json as Go struct.
//
// Takes in a VM location folder, similar to this: "/hast_shared/test-vm-1" or "/hast_shared/test-vm-1/" (trailing "/" automatically removed).
func GetVmConfig(vmLocation string) (r VmConfig, e error) {
	vmLocation = strings.TrimSuffix(vmLocation, "/")

	vmConfLocation := vmLocation + "/" + VM_CONFIG_NAME
	if !FileExists.CheckUsingOsStat(vmConfLocation) {
		e = errors.New("vm config file could not be found here: " + vmConfLocation)
		return
	}

	data, err := os.ReadFile(vmConfLocation)
	if err != nil {
		e = err
		return
	}

	err = json.Unmarshal(data, &r)
	if err != nil {
		e = err
		return
	}

	return
}
