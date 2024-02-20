// Copyright 2023 Hoster Authors. All rights reserved.
// Use of this source code is governed by an Apache License 2.0
// license that can be found in the LICENSE file.

package handlers

import (
	ApiAuth "HosterCore/internal/app/rest_api_v2/pkg/auth"
	ErrorMappings "HosterCore/internal/app/rest_api_v2/pkg/error_mappings"
	FreeBSDPgrep "HosterCore/internal/pkg/freebsd/pgrep"
	"HosterCore/internal/pkg/freebsd/rctl"
	HosterJailUtils "HosterCore/internal/pkg/hoster/jail/utils"
	"encoding/json"
	"net/http"
	"regexp"

	"github.com/gorilla/mux"
)

// @Tags Metrics, VMs
// @Summary Get the RCTL metrics for a specific VM.
// @Description Get the RCTL metrics for a specific VM.<br>`AUTH`: Only `rest` user is allowed.
// @Produce json
// @Security BasicAuth
// @Success 200 {object} rctl.RctMetrics
// @Failure 500 {object} SwaggerError
// @Param vm_name path string true "VM Name"
// @Router /metrics/vm/{vm_name} [get]
func VmMetrics(w http.ResponseWriter, r *http.Request) {
	if !ApiAuth.CheckRestUser(r) {
		user, pass, _ := r.BasicAuth()
		UnauthenticatedResponse(w, user, pass)
		return
	}

	vars := mux.Vars(r)
	vmName := vars["vm_name"]

	pids, err := FreeBSDPgrep.Pgrep(vmName)
	if err != nil {
		ReportError(w, http.StatusInternalServerError, err.Error())
		return
	}

	reMatchVm := regexp.MustCompile(`bhyve:\s+` + vmName + `$`)
	pid := -1
	for _, v := range pids {
		if reMatchVm.MatchString(v.ProcessCmd) {
			pid = v.ProcessId
		}
	}
	if pid < 0 {
		ReportError(w, http.StatusInternalServerError, ErrorMappings.VmIsNotRunning.String())
		return
	}

	rctl, err := rctl.MetricsProcess(pid)
	if err != nil {
		ReportError(w, http.StatusInternalServerError, err.Error())
		return
	}

	payload, err := json.Marshal(rctl)
	if err != nil {
		ReportError(w, http.StatusInternalServerError, err.Error())
		return
	}

	SetStatusCode(w, http.StatusOK)
	w.Write(payload)
}

// @Tags Metrics, Jails
// @Summary Get the RCTL metrics for a specific Jail.
// @Description Get the RCTL metrics for a specific Jail.<br>`AUTH`: Only `rest` user is allowed.
// @Produce json
// @Security BasicAuth
// @Success 200 {object} rctl.RctMetrics
// @Failure 500 {object} SwaggerError
// @Param jail_name path string true "Jail Name"
// @Router /metrics/jail/{jail_name} [get]
func JailMetrics(w http.ResponseWriter, r *http.Request) {
	if !ApiAuth.CheckRestUser(r) {
		user, pass, _ := r.BasicAuth()
		UnauthenticatedResponse(w, user, pass)
		return
	}

	vars := mux.Vars(r)
	jailName := vars["jail_name"]

	jlist, err := HosterJailUtils.GetRunningJails()
	if err != nil {
		ReportError(w, http.StatusInternalServerError, err.Error())
		return
	}

	jFound := false
	for _, v := range jlist {
		if v.Name == jailName {
			jFound = true
		}
	}
	if !jFound {
		ReportError(w, http.StatusInternalServerError, ErrorMappings.JailIsNotRunning.String())
		return
	}

	rctl, err := rctl.MetricsJail(jailName)
	if err != nil {
		ReportError(w, http.StatusInternalServerError, err.Error())
		return
	}

	payload, err := json.Marshal(rctl)
	if err != nil {
		ReportError(w, http.StatusInternalServerError, err.Error())
		return
	}

	SetStatusCode(w, http.StatusOK)
	w.Write(payload)
}
