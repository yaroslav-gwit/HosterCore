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
	HosterVm "HosterCore/internal/pkg/hoster/vm"
	HosterVmUtils "HosterCore/internal/pkg/hoster/vm/utils"
	"encoding/json"
	"net/http"

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
// @Router /vm/settings/{vm_name}/{new_tag} [post]
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

	err = HosterVmUtils.ConfigFileWriter(config, location)
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
