// Copyright 2024 Hoster Authors. All rights reserved.
// Use of this source code is governed by an Apache License 2.0
// license that can be found in the LICENSE file.

//go:build freebsd
// +build freebsd

package handlers

import (
	ApiAuth "HosterCore/internal/app/rest_api_v2/pkg/auth"
	JSONResponse "HosterCore/internal/app/rest_api_v2/pkg/json_response"
	HosterVm "HosterCore/internal/pkg/hoster/vm"
	HosterVmUtils "HosterCore/internal/pkg/hoster/vm/utils"
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
)

// @Tags VMs
// @Summary Destroy the VM.
// @Description Destroy the VM using it's name.<br>`AUTH`: Only `rest` user is allowed.
// @Produce json
// @Security BasicAuth
// @Success 200 {object} SwaggerSuccess
// @Failure 500 {object} SwaggerError
// @Param vm_name path string true "VM Name"
// @Router /vm/destroy/{vm_name} [delete]
func VmDestroy(w http.ResponseWriter, r *http.Request) {
	if !ApiAuth.CheckRestUser(r) {
		user, pass, _ := r.BasicAuth()
		UnauthenticatedResponse(w, user, pass)
		return
	}

	vars := mux.Vars(r)
	vmName := vars["vm_name"]

	err := HosterVm.Destroy(vmName)
	if err != nil {
		ReportError(w, http.StatusInternalServerError, err.Error())
		return
	}

	_, err = HosterVmUtils.WriteCache()
	if err != nil {
		ReportError(w, http.StatusInternalServerError, err.Error())
		return
	}

	payload, _ := JSONResponse.GenerateJson(w, "message", "success")
	SetStatusCode(w, http.StatusOK)
	w.Write(payload)
}

// @Tags VMs, Tags
// @Summary Delete an existing tag for any specific VM.
// @Description Delete an existing tag for any specific VM.<br>`AUTH`: Only `rest` user is allowed.
// @Produce json
// @Security BasicAuth
// @Success 200 {object} SwaggerSuccess
// @Failure 500 {object} SwaggerError
// @Param vm_name path string true "VM Name"
// @Param Input body TagInput true "Request payload"
// @Router /vm/settings/delete-tag/{vm_name} [delete]
func VmDeleteExistingTag(w http.ResponseWriter, r *http.Request) {
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

	tags := []string{}
	for _, v := range config.Tags {
		if v != input.Tag {
			tags = append(tags, v)
		}
	}

	config.Tags = []string{}
	config.Tags = append(config.Tags, tags...)

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
