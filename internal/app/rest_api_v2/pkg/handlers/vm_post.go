// Copyright 2024 Hoster Authors. All rights reserved.
// Use of this source code is governed by an Apache License 2.0
// license that can be found in the LICENSE file.

//go:build freebsd
// +build freebsd

package handlers

import (
	ApiAuth "HosterCore/internal/app/rest_api_v2/pkg/auth"
	ErrorMappings "HosterCore/internal/app/rest_api_v2/pkg/error_mappings"
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
// @Router /vm/settings/add-tag/{vm_name}/{new_tag} [post]
func VmPostNewTag(w http.ResponseWriter, r *http.Request) {
	if !ApiAuth.CheckRestUser(r) {
		user, pass, _ := r.BasicAuth()
		UnauthenticatedResponse(w, user, pass)
		return
	}

	vars := mux.Vars(r)
	newTag := vars["new_tag"]
	vmName := vars["vm_name"]

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
		if v == newTag {
			tagFound = true
			break
		}
	}
	if !tagFound {
		config.Tags = append(config.Tags, newTag)
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
func VmStart(w http.ResponseWriter, r *http.Request) {
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
// @Summary Stop a specific VM.
// @Description Stop a specific VM using it's name as a parameter.<br>`AUTH`: Both users are allowed.
// @Produce json
// @Security BasicAuth
// @Success 200 {object} SwaggerSuccess
// @Failure 500 {object} SwaggerError
// @Param Input body VmStopInput true "Request payload"
// @Router /vm/stop [post]
func VmStop(w http.ResponseWriter, r *http.Request) {
	if !ApiAuth.CheckAnyUser(r) {
		user, pass, _ := r.BasicAuth()
		UnauthenticatedResponse(w, user, pass)
		return
	}

	input := VmStopInput{}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&input)
	if err != nil {
		ReportError(w, http.StatusInternalServerError, ErrorMappings.CouldNotParseYourInput.String())
		return
	}

	err = HosterVm.Stop(input.VmName, false, false)
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
// @Router /vm/settings/firmware/{vm_name}/{production} [post]
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

	payload, _ := JSONResponse.GenerateJson(w, "message", "success")
	SetStatusCode(w, http.StatusOK)
	w.Write(payload)
}
