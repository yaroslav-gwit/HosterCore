// Copyright 2024 Hoster Authors. All rights reserved.
// Use of this source code is governed by an Apache License 2.0
// license that can be found in the LICENSE file.

//go:build freebsd
// +build freebsd

package handlers

import (
	ApiAuth "HosterCore/internal/app/rest_api_v2/pkg/auth"
	JSONResponse "HosterCore/internal/app/rest_api_v2/pkg/json_response"
	"HosterCore/internal/pkg/byteconversion"
	HosterHostUtils "HosterCore/internal/pkg/hoster/host/utils"
	HosterVm "HosterCore/internal/pkg/hoster/vm"
	HosterVmUtils "HosterCore/internal/pkg/hoster/vm/utils"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
)

// @Tags VMs, Tags
// @Summary Add a new tag for any particular VM.
// @Description Add a new tag for any particular VM.<br>`AUTH`: Only `rest` user is allowed.
// @Produce json
// @Security BasicAuth
// @Success 200 {object} SwaggerSuccess
// @Failure 500 {object} SwaggerError
// @Param vm_name path string true "VM Name"
// @Param new_tag path string true "New Tag"
// @Param Input body TagInput true "Request payload"
// @Router /vm/settings/add-tag/{vm_name} [post]
func VmPostNewTag(w http.ResponseWriter, r *http.Request) {
	if !ApiAuth.CheckRestUser(r) {
		user, pass, _ := r.BasicAuth()
		UnauthenticatedResponse(w, user, pass)
		return
	}

	vars := mux.Vars(r)
	vmName := vars["vm_name"]

	input := TagInput{}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&input)
	if err != nil {
		ReportError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if len(input.Tag) < 1 {
		ReportError(w, http.StatusBadRequest, "tag must be at least 1 character long")
		return
	}

	if len(input.Tag) > 255 {
		ReportError(w, http.StatusBadRequest, "tag must be at most 255 characters long")
		return
	}

	vmInfo, err := HosterVmUtils.InfoJsonApi(vmName)
	if err != nil {
		ReportError(w, http.StatusInternalServerError, err.Error())
		return
	}

	location := vmInfo.Simple.MountPoint.Mountpoint + "/" + vmName
	config, err := HosterVmUtils.GetVmConfig(location)
	if err != nil {
		ReportError(w, http.StatusInternalServerError, err.Error())
		return
	}

	var tagFound bool
	for _, v := range config.Tags {
		if v == input.Tag {
			tagFound = true
			break
		}
	}
	if !tagFound {
		config.Tags = append(config.Tags, input.Tag)
	}

	err = HosterVmUtils.ConfigFileWriter(config, location+"/"+HosterVmUtils.VM_CONFIG_NAME)
	if err != nil {
		ReportError(w, http.StatusInternalServerError, err.Error())
		return
	}

	payload, _ := JSONResponse.GenerateJson(w, "message", "success")
	SetStatusCode(w, http.StatusOK)
	w.Write(payload)

	SetStatusCode(w, http.StatusOK)
	w.Write(payload)
}

// @Tags VMs
// @Summary Deploy the new VM.
// @Description Deploy a new VM.<br>`AUTH`: Only `rest` user is allowed.
// @Produce json
// @Security BasicAuth
// @Success 200 {object} SwaggerSuccess
// @Failure 500 {object} SwaggerError
// @Param Input body HosterVm.VmDeployInput{} true "Request payload"
// @Router /vm/deploy [post]
func VmPostDeploy(w http.ResponseWriter, r *http.Request) {
	if !ApiAuth.CheckRestUser(r) {
		user, pass, _ := r.BasicAuth()
		UnauthenticatedResponse(w, user, pass)
		return
	}

	input := HosterVm.VmDeployInput{}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&input)
	if err != nil {
		ReportError(w, http.StatusInternalServerError, err.Error())
		return
	}

	err = HosterVm.Deploy(input)
	if err != nil {
		ReportError(w, http.StatusInternalServerError, err.Error())
		return
	}

	payload, _ := JSONResponse.GenerateJson(w, "message", "success")
	SetStatusCode(w, http.StatusOK)
	w.Write(payload)
}

// @Tags VMs
// @Summary Start a specific VM.
// @Description Start a specific VM using it's name as a parameter.<br>`AUTH`: Both users are allowed.
// @Produce json
// @Security BasicAuth
// @Success 200 {object} SwaggerSuccess
// @Failure 500 {object} SwaggerError
// @Param vm_name path string true "VM Name"
// @Router /vm/start/{vm_name} [post]
func VmPostStart(w http.ResponseWriter, r *http.Request) {
	if !ApiAuth.CheckAnyUser(r) {
		user, pass, _ := r.BasicAuth()
		UnauthenticatedResponse(w, user, pass)
		return
	}

	vars := mux.Vars(r)
	vmName := vars["vm_name"]

	err := HosterVm.Start(vmName, false, false)
	if err != nil {
		ReportError(w, http.StatusInternalServerError, err.Error())
		return
	}

	payload, _ := JSONResponse.GenerateJson(w, "message", "success")
	SetStatusCode(w, http.StatusOK)
	w.Write(payload)
}

// @Tags VMs
// @Summary Start all VMs.
// @Description Start all VMs.<br>`AUTH`: Both users are allowed.
// @Produce json
// @Security BasicAuth
// @Success 200 {object} SwaggerSuccess
// @Failure 500 {object} SwaggerError
// @Param production path bool true "Start only production VMs (true or false)"
// @Router /vm/start-all/{production} [post]
func VmPostStartAll(w http.ResponseWriter, r *http.Request) {
	if !ApiAuth.CheckAnyUser(r) {
		user, pass, _ := r.BasicAuth()
		UnauthenticatedResponse(w, user, pass)
		return
	}

	vars := mux.Vars(r)
	prodOnly := vars["production"]

	prod := false
	if strings.ToLower(prodOnly) == "true" {
		log.Debug("starting only production VMs")
		prod = true
	}

	go func(prod bool) {
		err := HosterVm.StartAll(prod, 1)
		if err != nil {
			// log.Errorf("Error starting all VMs: %s", err.Error())
			fmt.Printf("Error starting all VMs: %s\n", err.Error())
			return
		}
	}(prod)

	payload, _ := JSONResponse.GenerateJson(w, "message", "success")
	SetStatusCode(w, http.StatusOK)
	w.Write(payload)
}

// @Tags VMs
// @Summary Start a specific VM (and wait for a VNC screen connection).
// @Description Start a specific VM using it's name as a parameter (and wait for a VNC screen connection).<br>`AUTH`: Only `REST`-type user is allowed.
// @Produce json
// @Security BasicAuth
// @Success 200 {object} SwaggerSuccess
// @Failure 500 {object} SwaggerError
// @Param vm_name path string true "VM Name"
// @Router /vm/start/wait-vnc/{vm_name} [post]
func VmPostStartAndWaitVnc(w http.ResponseWriter, r *http.Request) {
	if !ApiAuth.CheckAnyUser(r) {
		user, pass, _ := r.BasicAuth()
		UnauthenticatedResponse(w, user, pass)
		return
	}

	vars := mux.Vars(r)
	vmName := vars["vm_name"]

	err := HosterVm.Start(vmName, true, false)
	if err != nil {
		ReportError(w, http.StatusInternalServerError, err.Error())
		return
	}

	payload, _ := JSONResponse.GenerateJson(w, "message", "success")
	SetStatusCode(w, http.StatusOK)
	w.Write(payload)
}

// @Tags VMs
// @Summary Clone the VM.
// @Description Clone the VM using it's name, and optionally specify the snapshot name to be used for cloning.<br>`AUTH`: Only `rest` user is allowed.
// @Produce json
// @Security BasicAuth
// @Success 200 {object} SwaggerSuccess
// @Failure 500 {object} SwaggerError
// @Param Input body VmCloneInput true "Request payload"
// @Router /vm/clone [post]
func VmClone(w http.ResponseWriter, r *http.Request) {
	if !ApiAuth.CheckRestUser(r) {
		user, pass, _ := r.BasicAuth()
		UnauthenticatedResponse(w, user, pass)
		return
	}

	input := VmCloneInput{}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&input)
	if err != nil {
		ReportError(w, http.StatusInternalServerError, err.Error())
		return
	}

	err = HosterVm.Clone(input.VmName, input.NewVmName, input.SnapshotName)
	if err != nil {
		ReportError(w, http.StatusInternalServerError, err.Error())
		return
	}

	payload, _ := JSONResponse.GenerateJson(w, "message", "success")
	SetStatusCode(w, http.StatusOK)
	w.Write(payload)
}

// @Tags VMs
// @Summary Modify VM's CPU settings.
// @Description Modify VM's CPU settings.<br>`AUTH`: Only `rest` user is allowed.
// @Produce json
// @Security BasicAuth
// @Success 200 {object} SwaggerSuccess
// @Failure 500 {object} SwaggerError
// @Param vm_name path string true "Name of the VM"
// @Param Input body VmCpuInput true "Request payload"
// @Router /vm/settings/cpu/{vm_name} [post]
func VmPostCpuInfo(w http.ResponseWriter, r *http.Request) {
	if !ApiAuth.CheckRestUser(r) {
		user, pass, _ := r.BasicAuth()
		UnauthenticatedResponse(w, user, pass)
		return
	}

	vars := mux.Vars(r)
	vmName := vars["vm_name"]

	input := VmCpuInput{}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&input)
	if err != nil {
		ReportError(w, http.StatusInternalServerError, err.Error())
		return
	}

	hostInfo, err := HosterHostUtils.GetHostInfo()
	if err != nil {
		ReportError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if input.CpuCores < 1 {
		ReportError(w, http.StatusInternalServerError, "CPU cores must be greater than 0")
		return
	}
	if input.CpuThreads < 1 {
		ReportError(w, http.StatusInternalServerError, "CPU threads must be greater than 0")
		return
	}
	if input.CpuSockets < 1 {
		ReportError(w, http.StatusInternalServerError, "CPU sockets must be greater than 0")
		return
	}

	overallCpus := input.CpuCores * input.CpuThreads * input.CpuSockets
	if overallCpus > hostInfo.CpuInfo.OverallCpus || input.CpuCores*input.CpuThreads < 1 {
		ReportError(w, http.StatusInternalServerError, "CPU settings exceed the host's capabilities")
		return
	}

	vmInfo, err := HosterVmUtils.InfoJsonApi(vmName)
	if err != nil {
		ReportError(w, http.StatusInternalServerError, err.Error())
		return
	}

	location := vmInfo.Simple.MountPoint.Mountpoint + "/" + vmName
	config, err := HosterVmUtils.GetVmConfig(location)
	if err != nil {
		ReportError(w, http.StatusInternalServerError, err.Error())
		return
	}

	config.CPUCores = input.CpuCores
	config.CPUThreads = input.CpuThreads
	config.CPUSockets = input.CpuSockets

	err = HosterVmUtils.ConfigFileWriter(config, location+"/"+HosterVmUtils.VM_CONFIG_NAME)
	if err != nil {
		ReportError(w, http.StatusInternalServerError, err.Error())
		return
	}

	payload, _ := JSONResponse.GenerateJson(w, "message", "success")
	SetStatusCode(w, http.StatusOK)
	w.Write(payload)
}

// @Tags VMs
// @Summary Modify VM's RAM settings.
// @Description Modify VM's RAM settings.<br>`AUTH`: Only `rest` user is allowed.
// @Produce json
// @Security BasicAuth
// @Success 200 {object} SwaggerSuccess
// @Failure 500 {object} SwaggerError
// @Param vm_name path string true "Name of the VM"
// @Param Input body VmRamInput true "Request payload"
// @Router /vm/settings/ram/{vm_name} [post]
func VmPostRamInfo(w http.ResponseWriter, r *http.Request) {
	if !ApiAuth.CheckRestUser(r) {
		user, pass, _ := r.BasicAuth()
		UnauthenticatedResponse(w, user, pass)
		return
	}

	vars := mux.Vars(r)
	vmName := vars["vm_name"]

	input := VmRamInput{}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&input)
	if err != nil {
		ReportError(w, http.StatusInternalServerError, err.Error())
		return
	}

	hostInfo, err := HosterHostUtils.GetHostInfo()
	if err != nil {
		ReportError(w, http.StatusInternalServerError, err.Error())
		return
	}

	overallRamHuman := fmt.Sprintf("%d%s", input.RamAmount, input.BytesValue)
	overallRamBytes, err := byteconversion.HumanToBytes(overallRamHuman)
	if err != nil {
		ReportError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if input.BytesValue == "M" {
		if input.RamAmount < 512 {
			ReportError(w, http.StatusInternalServerError, "RAM must be at least 512MB")
			return
		}
	} else if input.BytesValue == "G" {
		if input.RamAmount < 1 {
			ReportError(w, http.StatusInternalServerError, "RAM must be at least 1GB")
			return
		}
	} else {
		ReportError(w, http.StatusInternalServerError, "Invalid RAM value")
		return
	}

	if overallRamBytes >= hostInfo.RamInfo.RamOverallBytes {
		ReportError(w, http.StatusInternalServerError, "RAM settings exceed the host's capabilities")
		return
	}

	vmInfo, err := HosterVmUtils.InfoJsonApi(vmName)
	if err != nil {
		ReportError(w, http.StatusInternalServerError, err.Error())
		return
	}

	location := vmInfo.Simple.MountPoint.Mountpoint + "/" + vmName
	config, err := HosterVmUtils.GetVmConfig(location)
	if err != nil {
		ReportError(w, http.StatusInternalServerError, err.Error())
		return
	}

	config.Memory = overallRamHuman
	// log.Debug("Setting RAM to: " + overallRamHuman)
	// log.Debug(config)

	err = HosterVmUtils.ConfigFileWriter(config, location+"/"+HosterVmUtils.VM_CONFIG_NAME)
	if err != nil {
		ReportError(w, http.StatusInternalServerError, err.Error())
		return
	}

	payload, _ := JSONResponse.GenerateJson(w, "message", "success")
	SetStatusCode(w, http.StatusOK)
	w.Write(payload)
}

// @Tags VMs
// @Summary Modify VM's VNC Resolution.
// @Description Modify VM's VNC Resolution.<br>`AUTH`: Only `rest` user is allowed.
// @Produce json
// @Security BasicAuth
// @Success 200 {object} SwaggerSuccess
// @Failure 500 {object} SwaggerError
// @Param vm_name path string true "Name of the VM"
// @Param resolution path int true "Resolution code, e.g. 3 for 1024x768"
// @Router /vm/settings/vnc-resolution/{vm_name}/{resolution} [post]
func VmPostVncResolution(w http.ResponseWriter, r *http.Request) {
	if !ApiAuth.CheckRestUser(r) {
		user, pass, _ := r.BasicAuth()
		UnauthenticatedResponse(w, user, pass)
		return
	}

	vars := mux.Vars(r)
	vmName := vars["vm_name"]
	resolution := vars["resolution"]

	resolutionInt, err := strconv.Atoi(resolution)
	if err != nil {
		ReportError(w, http.StatusInternalServerError, "resolution must be an integer")
		return
	}

	if resolutionInt < 1 || resolutionInt > 9 {
		// VNC Resolution List
		// 9: 1920x1200
		// 8: 1920x1080
		// 7: 1600x1200
		// 6: 1600x900
		// 5: 1280x1024
		// 4: 1280x720
		// 3: 1024x768
		// 2: 800x600
		// 1: 640x480
		ReportError(w, http.StatusInternalServerError, "invalid screen resolution")
		return
	}

	vmInfo, err := HosterVmUtils.InfoJsonApi(vmName)
	if err != nil {
		ReportError(w, http.StatusInternalServerError, err.Error())
		return
	}

	location := vmInfo.Simple.MountPoint.Mountpoint + "/" + vmName
	config, err := HosterVmUtils.GetVmConfig(location)
	if err != nil {
		ReportError(w, http.StatusInternalServerError, err.Error())
		return
	}

	config.VncResolution = resolutionInt
	err = HosterVmUtils.ConfigFileWriter(config, location+"/"+HosterVmUtils.VM_CONFIG_NAME)
	if err != nil {
		ReportError(w, http.StatusInternalServerError, err.Error())
		return
	}

	payload, _ := JSONResponse.GenerateJson(w, "message", "success")
	SetStatusCode(w, http.StatusOK)
	w.Write(payload)
}

// @Tags VMs
// @Summary Modify VM's Firmware type (e.g. bootloader type, bios vs uefi).
// @Description Modify VM's Firmware type (e.g. bootloader type, bios vs uefi).<br>`AUTH`: Only `rest` user is allowed.
// @Produce json
// @Security BasicAuth
// @Success 200 {object} SwaggerSuccess
// @Failure 500 {object} SwaggerError
// @Param vm_name path string true "Name of the VM"
// @Param firmware path string true "Firmware type (bootloader type), e.g. bios or uefi"
// @Router /vm/settings/firmware/{vm_name}/{firmware} [post]
func VmPostFirmwareType(w http.ResponseWriter, r *http.Request) {
	if !ApiAuth.CheckRestUser(r) {
		user, pass, _ := r.BasicAuth()
		UnauthenticatedResponse(w, user, pass)
		return
	}

	vars := mux.Vars(r)
	vmName := vars["vm_name"]
	firmware := vars["firmware"]

	if firmware == "bios" || firmware == "uefi" {
		_ = 0
	} else {
		errValue := "invalid firmware type, must be either 'bios' or 'uefi'"
		ReportError(w, http.StatusInternalServerError, errValue)
		return
	}

	vmInfo, err := HosterVmUtils.InfoJsonApi(vmName)
	if err != nil {
		ReportError(w, http.StatusInternalServerError, err.Error())
		return
	}

	location := vmInfo.Simple.MountPoint.Mountpoint + "/" + vmName
	config, err := HosterVmUtils.GetVmConfig(location)
	if err != nil {
		ReportError(w, http.StatusInternalServerError, err.Error())
		return
	}

	config.Loader = firmware
	err = HosterVmUtils.ConfigFileWriter(config, location+"/"+HosterVmUtils.VM_CONFIG_NAME)
	if err != nil {
		ReportError(w, http.StatusInternalServerError, err.Error())
		return
	}

	payload, _ := JSONResponse.GenerateJson(w, "message", "success")
	SetStatusCode(w, http.StatusOK)
	w.Write(payload)
}

// @Tags VMs
// @Summary Modify VM's Workload type (e.g. is this a production VM, true or false).
// @Description Modify VM's Workload type (e.g. is this a production VM, true or false).<br>`AUTH`: Only `rest` user is allowed.
// @Produce json
// @Security BasicAuth
// @Success 200 {object} SwaggerSuccess
// @Failure 500 {object} SwaggerError
// @Param vm_name path string true "Name of the VM"
// @Param production path string true "Workload type (is this a production VM), e.g. true or false"
// @Router /vm/settings/production/{vm_name}/{production} [post]
func VmPostProductionSetting(w http.ResponseWriter, r *http.Request) {
	if !ApiAuth.CheckRestUser(r) {
		user, pass, _ := r.BasicAuth()
		UnauthenticatedResponse(w, user, pass)
		return
	}

	vars := mux.Vars(r)
	vmName := vars["vm_name"]
	prod := vars["production"]

	prod = strings.ToLower(prod)
	if prod == "true" || prod == "false" {
		_ = 0
	} else {
		errValue := "invalid workload type, must be either 'true' or 'false'"
		ReportError(w, http.StatusInternalServerError, errValue)
		return
	}

	vmInfo, err := HosterVmUtils.InfoJsonApi(vmName)
	if err != nil {
		ReportError(w, http.StatusInternalServerError, err.Error())
		return
	}

	location := vmInfo.Simple.MountPoint.Mountpoint + "/" + vmName
	config, err := HosterVmUtils.GetVmConfig(location)
	if err != nil {
		ReportError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if prod == "true" {
		config.Production = true
	} else {
		config.Production = false
	}
	err = HosterVmUtils.ConfigFileWriter(config, location+"/"+HosterVmUtils.VM_CONFIG_NAME)
	if err != nil {
		ReportError(w, http.StatusInternalServerError, err.Error())
		return
	}

	payload, _ := JSONResponse.GenerateJson(w, "message", "success")
	SetStatusCode(w, http.StatusOK)
	w.Write(payload)
}

// @Tags VMs
// @Summary Modify VM's OS info (e.g. os_type - debian12, os_comment - Debian 12).
// @Description Modify VM's OS info (e.g. os_type - debian12, os_comment - Debian 12).<br>`AUTH`: Only `rest` user is allowed.
// @Produce json
// @Security BasicAuth
// @Success 200 {object} SwaggerSuccess
// @Failure 500 {object} SwaggerError
// @Param vm_name path string true "Name of the VM"
// @Param Input body VmOsSettings true "Request payload"
// @Router /vm/settings/os-info/{vm_name} [post]
func VmPostOsSettings(w http.ResponseWriter, r *http.Request) {
	if !ApiAuth.CheckRestUser(r) {
		user, pass, _ := r.BasicAuth()
		UnauthenticatedResponse(w, user, pass)
		return
	}

	vars := mux.Vars(r)
	vmName := vars["vm_name"]

	input := VmOsSettings{}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&input)
	if err != nil {
		ReportError(w, http.StatusInternalServerError, err.Error())
		return
	}

	vmInfo, err := HosterVmUtils.InfoJsonApi(vmName)
	if err != nil {
		ReportError(w, http.StatusInternalServerError, err.Error())
		return
	}

	location := vmInfo.Simple.MountPoint.Mountpoint + "/" + vmName
	config, err := HosterVmUtils.GetVmConfig(location)
	if err != nil {
		ReportError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Debug
	// log.Debug(input)

	config.OsType = input.OsType
	config.OsComment = input.OsComment

	err = HosterVmUtils.ConfigFileWriter(config, location+"/"+HosterVmUtils.VM_CONFIG_NAME)
	if err != nil {
		ReportError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Update the VM's cache because we've changed the VM's settings
	// (otherwise the icon won't change in the UI)
	_, err = HosterVmUtils.WriteCache()
	if err != nil {
		ReportError(w, http.StatusInternalServerError, err.Error())
		return
	}

	payload, _ := JSONResponse.GenerateJson(w, "message", "success")
	SetStatusCode(w, http.StatusOK)
	w.Write(payload)
}

// @Tags VMs
// @Summary Stop a specific VM.
// @Description Stop a specific VM using it's name as a parameter.<br>`AUTH`: Both users are allowed.
// @Produce json
// @Security BasicAuth
// @Success 200 {object} SwaggerSuccess
// @Failure 500 {object} SwaggerError
// @Param vm_name path string true "VM Name"
// @Router /vm/stop/{vm_name} [post]
func VmPostStop(w http.ResponseWriter, r *http.Request) {
	if !ApiAuth.CheckAnyUser(r) {
		user, pass, _ := r.BasicAuth()
		UnauthenticatedResponse(w, user, pass)
		return
	}

	vars := mux.Vars(r)
	vmName := vars["vm_name"]

	err := HosterVm.Stop(vmName, false, false)
	if err != nil {
		ReportError(w, http.StatusInternalServerError, err.Error())
		return
	}

	payload, _ := JSONResponse.GenerateJson(w, "message", "success")
	SetStatusCode(w, http.StatusOK)
	w.Write(payload)
}

// @Tags VMs
// @Summary Stop (forcefully) a specific VM.
// @Description Stop (forcefully) a specific VM using it's name as a parameter.<br>`AUTH`: Both users are allowed.
// @Produce json
// @Security BasicAuth
// @Success 200 {object} SwaggerSuccess
// @Failure 500 {object} SwaggerError
// @Param vm_name path string true "VM Name"
// @Router /vm/stop/force/{vm_name} [post]
func VmPostStopForce(w http.ResponseWriter, r *http.Request) {
	if !ApiAuth.CheckAnyUser(r) {
		user, pass, _ := r.BasicAuth()
		UnauthenticatedResponse(w, user, pass)
		return
	}

	vars := mux.Vars(r)
	vmName := vars["vm_name"]

	err := HosterVm.Stop(vmName, true, false)
	if err != nil {
		ReportError(w, http.StatusInternalServerError, err.Error())
		return
	}

	payload, _ := JSONResponse.GenerateJson(w, "message", "success")
	SetStatusCode(w, http.StatusOK)
	w.Write(payload)
}

// @Tags VMs
// @Summary Replace a real CloudInit ISO with an empty one.
// @Description Replace a real CloudInit ISO with an empty one. Useful in the situations where multiple users reside on the same VM, because an empty ISO will protect the VM's secrets.<br>`AUTH`: Only `REST` user is allowed.
// @Produce json
// @Security BasicAuth
// @Success 200 {object} SwaggerSuccess
// @Failure 500 {object} SwaggerError
// @Param vm_name path string true "VM Name"
// @Router /vm/cloud-init/unmount-iso/{vm_name} [post]
func VmPostUnmountCiIso(w http.ResponseWriter, r *http.Request) {
	if !ApiAuth.CheckRestUser(r) {
		user, pass, _ := r.BasicAuth()
		UnauthenticatedResponse(w, user, pass)
		return
	}

	vars := mux.Vars(r)
	vmName := vars["vm_name"]

	err := HosterVmUtils.UnmountCiIso(vmName)
	if err != nil {
		ReportError(w, http.StatusInternalServerError, err.Error())
		return
	}

	payload, _ := JSONResponse.GenerateJson(w, "message", "success")
	SetStatusCode(w, http.StatusOK)
	w.Write(payload)
}

// @Tags VMs
// @Summary Mount a real CloudInit ISO.
// @Description Mount a real CloudInit ISO.<br>`AUTH`: Only `REST` user is allowed.
// @Produce json
// @Security BasicAuth
// @Success 200 {object} SwaggerSuccess
// @Failure 500 {object} SwaggerError
// @Param vm_name path string true "VM Name"
// @Router /vm/cloud-init/mount-iso/{vm_name} [post]
func VmPostMountCiIso(w http.ResponseWriter, r *http.Request) {
	if !ApiAuth.CheckRestUser(r) {
		user, pass, _ := r.BasicAuth()
		UnauthenticatedResponse(w, user, pass)
		return
	}

	vars := mux.Vars(r)
	vmName := vars["vm_name"]

	err := HosterVmUtils.MountCiIso(vmName)
	if err != nil {
		ReportError(w, http.StatusInternalServerError, err.Error())
		return
	}

	payload, _ := JSONResponse.GenerateJson(w, "message", "success")
	SetStatusCode(w, http.StatusOK)
	w.Write(payload)
}

// @Tags VMs
// @Summary Mount a real ISO.
// @Description Mount a real ISO. This could be an installation ISO, or an ISO with OS drivers, etc.<br>`AUTH`: Only `REST` user is allowed.
// @Produce json
// @Security BasicAuth
// @Success 200 {object} SwaggerSuccess
// @Failure 500 {object} SwaggerError
// @Param vm_name path string true "VM Name"
// @Router /vm/settings/mount-iso/{vm_name} [post]
func VmPostMountIso(w http.ResponseWriter, r *http.Request) {
	if !ApiAuth.CheckRestUser(r) {
		user, pass, _ := r.BasicAuth()
		UnauthenticatedResponse(w, user, pass)
		return
	}

	vars := mux.Vars(r)
	vmName := vars["vm_name"]

	input := VmMountIsoInput{}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&input)
	if err != nil {
		ReportError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// log.Debug(input)
	err = HosterVmUtils.MountInstallationIso(vmName, input.IsoPath, input.IsoComment)
	if err != nil {
		ReportError(w, http.StatusInternalServerError, err.Error())
		return
	}

	payload, _ := JSONResponse.GenerateJson(w, "message", "success")
	SetStatusCode(w, http.StatusOK)
	w.Write(payload)
}

// @Tags VMs
// @Summary Unmount an installation ISO.
// @Description Unmount an installation ISO. This could be an installation ISO, or an ISO with OS drivers, etc.<br>`AUTH`: Only `REST` user is allowed.
// @Produce json
// @Security BasicAuth
// @Success 200 {object} SwaggerSuccess
// @Failure 500 {object} SwaggerError
// @Param vm_name path string true "VM Name"
// @Router /vm/settings/unmount-iso/{vm_name} [post]
func VmPostUnmountIso(w http.ResponseWriter, r *http.Request) {
	if !ApiAuth.CheckRestUser(r) {
		user, pass, _ := r.BasicAuth()
		UnauthenticatedResponse(w, user, pass)
		return
	}

	vars := mux.Vars(r)
	vmName := vars["vm_name"]

	input := VmMountIsoInput{}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&input)
	if err != nil {
		ReportError(w, http.StatusInternalServerError, err.Error())
		return
	}

	err = HosterVmUtils.UnmountInstallationIso(vmName, input.IsoPath)
	if err != nil {
		ReportError(w, http.StatusInternalServerError, err.Error())
		return
	}

	payload, _ := JSONResponse.GenerateJson(w, "message", "success")
	SetStatusCode(w, http.StatusOK)
	w.Write(payload)
}

// @Tags VMs
// @Summary Add a new VM data disk.
// @Description Add a new VM data disk.<br>`AUTH`: Only `rest` user is allowed.
// @Produce json
// @Security BasicAuth
// @Success 200 {object} SwaggerSuccess
// @Failure 500 {object} SwaggerError
// @Param vm_name path string true "Name of the VM"
// @Param Input body HosterVmUtils.VmDisk{} true "Request payload"
// @Router /vm/settings/disk/add-new/{vm_name} [post]
func VmPostAddNewDisk(w http.ResponseWriter, r *http.Request) {
	if !ApiAuth.CheckRestUser(r) {
		user, pass, _ := r.BasicAuth()
		UnauthenticatedResponse(w, user, pass)
		return
	}

	vars := mux.Vars(r)
	vmName := vars["vm_name"]

	input := HosterVmUtils.VmDisk{}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&input)
	if err != nil {
		ReportError(w, http.StatusInternalServerError, err.Error())
		return
	}

	err = HosterVmUtils.AddNewVmDisk(vmName, input)
	if err != nil {
		ReportError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Update the VM's cache because we've changed the VM's settings
	// (otherwise the icon won't change in the UI)
	// _, err = HosterVmUtils.WriteCache()
	// if err != nil {
	// 	ReportError(w, http.StatusInternalServerError, err.Error())
	// 	return
	// }

	payload, _ := JSONResponse.GenerateJson(w, "message", "success")
	SetStatusCode(w, http.StatusOK)
	w.Write(payload)
}

// @Tags VMs
// @Summary Expand an existing VM disk.
// @Description Expand an existing VM disk.<br>`AUTH`: Only `rest` user is allowed.
// @Produce json
// @Security BasicAuth
// @Success 200 {object} SwaggerSuccess
// @Failure 500 {object} SwaggerError
// @Param vm_name path string true "Name of the VM"
// @Param Input body VmDiskExpandInput{} true "Request payload"
// @Router /vm/settings/disk/expand/{vm_name} [post]
func VmPostExpandDisk(w http.ResponseWriter, r *http.Request) {
	if !ApiAuth.CheckRestUser(r) {
		user, pass, _ := r.BasicAuth()
		UnauthenticatedResponse(w, user, pass)
		return
	}

	vars := mux.Vars(r)
	vmName := vars["vm_name"]

	input := VmDiskExpandInput{}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&input)
	if err != nil {
		ReportError(w, http.StatusInternalServerError, err.Error())
		return
	}

	err = HosterVmUtils.DiskExpandOffline(input.DiskImage, input.ExpansionSize, vmName)
	if err != nil {
		ReportError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Update the VM's cache because we've changed the VM's settings
	// (otherwise the icon won't change in the UI)
	// _, err = HosterVmUtils.WriteCache()
	// if err != nil {
	// 	ReportError(w, http.StatusInternalServerError, err.Error())
	// 	return
	// }

	payload, _ := JSONResponse.GenerateJson(w, "message", "success")
	SetStatusCode(w, http.StatusOK)
	w.Write(payload)
}

// @Tags VMs, Networks
// @Summary Add a new VM network interface.
// @Description Add a new VM network interface.<br>`AUTH`: Only `rest` user is allowed.
// @Produce json
// @Security BasicAuth
// @Success 200 {object} SwaggerSuccess
// @Failure 500 {object} SwaggerError
// @Param vm_name path string true "Name of the VM"
// @Param Input body HosterVmUtils.VmNetwork{} true "Request payload"
// @Router /vm/settings/network/add/{vm_name} [post]
func VmPostAddNewNetwork(w http.ResponseWriter, r *http.Request) {
	if !ApiAuth.CheckRestUser(r) {
		user, pass, _ := r.BasicAuth()
		UnauthenticatedResponse(w, user, pass)
		return
	}

	vars := mux.Vars(r)
	vmName := vars["vm_name"]

	input := HosterVmUtils.VmNetwork{}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&input)
	if err != nil {
		ReportError(w, http.StatusInternalServerError, err.Error())
		return
	}

	err = HosterVm.AddNewVmNetwork(vmName, input)
	if err != nil {
		ReportError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Update the VM's cache because we've changed the VM's settings
	// (otherwise the icon won't change in the UI)
	// _, err = HosterVmUtils.WriteCache()
	// if err != nil {
	// 	ReportError(w, http.StatusInternalServerError, err.Error())
	// 	return
	// }

	payload, _ := JSONResponse.GenerateJson(w, "message", "success")
	SetStatusCode(w, http.StatusOK)
	w.Write(payload)
}

// @Tags VMs
// @Summary Update VM's description.
// @Description Update VM's description.<br>`AUTH`: Only `rest` user is allowed.
// @Produce json
// @Security BasicAuth
// @Success 200 {object} SwaggerSuccess
// @Failure 500 {object} SwaggerError
// @Param vm_name path string true "Name of the VM"
// @Param Input body ResourceDescription{} true "Request payload"
// @Router /vm/settings/description/{vm_name} [post]
func VmPostDescription(w http.ResponseWriter, r *http.Request) {
	if !ApiAuth.CheckRestUser(r) {
		user, pass, _ := r.BasicAuth()
		UnauthenticatedResponse(w, user, pass)
		return
	}

	vars := mux.Vars(r)
	vmName := vars["vm_name"]

	input := ResourceDescription{}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&input)
	if err != nil {
		ReportError(w, http.StatusInternalServerError, err.Error())
		return
	}

	err = HosterVmUtils.UpdateDescription(vmName, input.Description)
	if err != nil {
		ReportError(w, http.StatusInternalServerError, err.Error())
		return
	}

	payload, _ := JSONResponse.GenerateJson(w, "message", "success")
	SetStatusCode(w, http.StatusOK)
	w.Write(payload)
}
