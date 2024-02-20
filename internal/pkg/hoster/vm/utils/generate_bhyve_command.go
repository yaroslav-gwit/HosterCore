// Copyright 2024 Hoster Authors. All rights reserved.
// Use of this source code is governed by an Apache License 2.0
// license that can be found in the LICENSE file.

package HosterVmUtils

import (
	"HosterCore/internal/pkg/emojlog"
	HosterNetwork "HosterCore/internal/pkg/hoster/network"
	"fmt"
	"slices"
	"strconv"
	"strings"
)

// This function generates a bhyve start command, that is then passed down to the VM Supervisor.
func GenerateBhyveStartCmd(vmName string, vmLocation string, restoreVmState bool, waitVnc bool) (r string, e error) {
	vmLocation = strings.TrimSuffix(vmLocation, "/")
	conf, err := GetVmConfig(vmLocation)
	if err != nil {
		e = err
		return
	}

	if len(conf.Passthru) > 0 {
		r = "bhyve -S -HAuw -s 0:0,hostbridge -s 31,lpc "
	} else {
		r = "bhyve -HAuw -s 0:0,hostbridge -s 31,lpc "
	}
	bhyvePci := 2
	bhyvePci2 := 0

	// Generate network config
	var networkFinal string
	for _, v := range conf.Networks {
		tap, err := HosterNetwork.CreateTapInterface(vmName, v.NetworkBridge)
		if err != nil {
			e = err
			return
		}
		if bhyvePci == 2 {
			_ = 0
		} else {
			bhyvePci += 1
		}
		networkFinal = fmt.Sprintf("%s -s %d:%d,%s,%s,mac=%s", networkFinal, bhyvePci, bhyvePci2, v.NetworkAdaptorType, tap, v.NetworkMac)
	}
	r = r + networkFinal
	// EOF Generate network config

	// Generate disk config
	var diskFinal string
	for _, v := range conf.Disks {
		var diskImageLocation string
		if v.DiskLocation == "internal" {
			diskImageLocation = vmLocation + "/" + v.DiskImage
		} else {
			diskImageLocation = v.DiskImage
		}
		bhyvePci = bhyvePci + 1
		diskFinal = fmt.Sprintf("%s -s %d:%d,%s,%s", diskFinal, bhyvePci, bhyvePci2, v.DiskType, diskImageLocation)
	}
	r = r + diskFinal
	// EOF Generate disk config

	// VirtIO 9P
	if len(conf.Shares) > 0 {
		share9Pcommand := ""
		for _, v := range conf.Shares {
			bhyvePci = bhyvePci + 1
			if v.ReadOnly {
				share9Pcommand = fmt.Sprintf("%s -s %d,virtio-9p,%s=%s,ro", share9Pcommand, bhyvePci, v.ShareName, v.ShareLocation)
			} else {
				share9Pcommand = fmt.Sprintf("%s -s %d,virtio-9p,%s=%s", share9Pcommand, bhyvePci, v.ShareName, v.ShareLocation)
			}
		}
		r = r + share9Pcommand
	}
	// EOF VirtIO 9P

	// Generate CPU and RAM config
	var cpuAndRam string
	if conf.CPUThreads > 0 {
		cpuAndRam = fmt.Sprintf(" -c sockets=%d,cores=%d,threads=%d -m %s", conf.CPUSockets, conf.CPUCores, conf.CPUThreads, conf.Memory)
	} else {
		cpuAndRam = fmt.Sprintf(" -c sockets=%d,cores=%d -m %s", conf.CPUSockets, conf.CPUCores, conf.Memory)
	}
	r = r + cpuAndRam
	// fmt.Println(bhyveFinalCommand)
	// EOF Generate CPU and RAM config

	// VNC options
	bhyvePci = bhyvePci + 1
	bhyvePci2 = 0
	vncResolution := setScreenResolution(conf.VncResolution)
	vncCommand := fmt.Sprintf(" -s %d:%d,fbuf,tcp=0.0.0.0:%d,%s,password=%s", bhyvePci, bhyvePci2, conf.VncPort, vncResolution, conf.VncPassword)
	// Set the VGA mode if found in the config file
	if len(conf.VGA) > 0 {
		if conf.VGA == "io" || conf.VGA == "on" || conf.VGA == "off" {
			_ = 0
		}
		vncCommand = vncCommand + ",vga=" + conf.VGA
	}
	if waitVnc { // Wait for the VNC connection before booting the VM, if waitVnc was enabled at runtime
		vncCommand = vncCommand + ",wait"
	}
	r = r + vncCommand
	// fmt.Println(bhyveFinalCommand)
	// EOF VNC options

	// Passthru section
	passthruString := ""
	if len(conf.Passthru) > 0 {
		// PCI slot added for each card iteration, no need to add one before the loop
		bhyvePci = bhyvePci + 1
		passthruString, bhyvePci = passthruPciSplitter(bhyvePci, conf.Passthru)
		r = r + " " + passthruString
	}
	// EOF Passthru section

	// XHCI and LOADER section
	bhyvePci = bhyvePci + 1
	bhyvePci2 = 0

	var xhciCommand string
	if !conf.DisableXHCI {
		xhciCommand = fmt.Sprintf(" -s %d:%d,xhci,tablet", bhyvePci, bhyvePci2)
	}

	var loaderCommand string
	if conf.Loader == "bios" {
		loaderCommand = fmt.Sprintf("%s -l com1,/dev/nmdm-%s-1A -l bootrom,/usr/local/share/uefi-firmware/BHYVE_UEFI_CSM.fd", xhciCommand, vmName)
	} else if conf.Loader == "uefi" {
		loaderCommand = fmt.Sprintf("%s -l com1,/dev/nmdm-%s-1A -l bootrom,/usr/local/share/uefi-firmware/BHYVE_UEFI.fd", xhciCommand, vmName)
	}
	// EOF XHCI and LOADER section

	if len(conf.UUID) > 0 {
		loaderCommand = loaderCommand + " -U " + conf.UUID
	}

	if restoreVmState {
		loaderCommand = loaderCommand + " -r " + vmLocation + "/vm_state"
	} else {
		loaderCommand = loaderCommand + " -u " + vmName
	}

	for strings.Contains(r, "  ") {
		r = strings.ReplaceAll(r, "  ", " ")
	}

	r = r + loaderCommand
	return
}

// Takes in a list of PCI devices like so: [ "4/0/0", "43/0/1", "43/0/12" ]
//
// And returns the correct mappings for the Bhyve passthru, like so:
// -s 6:0,passthru,4/0/0 -s 7:1,passthru,43/0/1 -s 7:12,passthru,43/0/12
func passthruPciSplitter(startWithId int, devices []string) (pciDevs string, latestPciId int) {
	group0 := ""
	group1 := ""
	iter := 0
	skipThisIterationList := []int{}

	for i, v := range devices {
		if len(strings.Split(v, "/")) < 3 {
			emojlog.PrintLogMessage("This PCI device would not be added to passthru: "+v+", because it uses the incorrect format", emojlog.Error)
			continue
		}
		if slices.Contains(skipThisIterationList, i) {
			continue
		}

		group0 = strings.Split(v, "/")[0]
		group1 = strings.Split(v, "/")[1]
		iter = i

		if strings.HasPrefix(v, "-") {
			vNoPrefix := strings.TrimPrefix(v, "-")
			pciDevs = pciDevs + " -s " + strconv.Itoa(startWithId) + ":0,passthru," + vNoPrefix
		} else {
			pciDevs = pciDevs + " -s " + strconv.Itoa(startWithId) + ":" + strings.Split(v, "/")[2] + ",passthru," + v
		}
		for ii, vv := range devices {
			if ii == iter || strings.HasPrefix(vv, "-") || strings.HasPrefix(v, "-") {
				continue
			}
			if slices.Contains(skipThisIterationList, ii) {
				continue
			}
			if len(strings.Split(vv, "/")) < 3 {
				emojlog.PrintLogMessage("This PCI device would not be added to passthru: "+vv+", because it uses the incorrect format", emojlog.Error)
				continue
			}

			vvSpit := strings.Split(vv, "/")
			if strings.TrimSpace(vvSpit[0]) == group0 && strings.TrimSpace(vvSpit[1]) == group1 {
				skipThisIterationList = append(skipThisIterationList, ii)
				pciDevs = pciDevs + " -s " + strconv.Itoa(startWithId) + ":" + strings.Split(vv, "/")[2] + ",passthru," + vv
			}
		}
		startWithId += 1
	}

	pciDevs = strings.TrimSpace(pciDevs)
	if iter > 0 {
		latestPciId = startWithId - 1
	}
	return
}

// Generate a VNC resolution from a pre-set integer.
func setScreenResolution(input int) (screenRes string) {
	// default case
	screenRes = "w=800,h=600"

	// case switch
	if input == 1 {
		screenRes = "w=640,h=480"
	} else if input == 2 {
		screenRes = "w=800,h=600"
	} else if input == 3 {
		screenRes = "w=1024,h=768"
	} else if input == 4 {
		screenRes = "w=1280,h=720"
	} else if input == 5 {
		screenRes = "w=1280,h=1024"
	} else if input == 6 {
		screenRes = "w=1600,h=900"
	} else if input == 7 {
		screenRes = "w=1600,h=1200"
	} else if input == 8 {
		screenRes = "w=1920,h=1080"
	} else if input == 9 {
		screenRes = "w=1920,h=1200"
	}

	return
}
