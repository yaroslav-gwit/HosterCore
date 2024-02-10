// Copyright 2023 Hoster Authors. All rights reserved.
// Use of this source code is governed by an Apache License 2.0
// license that can be found in the LICENSE file.

package handlers

import (
	ApiAuth "HosterCore/internal/app/rest_api_v2/pkg/auth"
	ErrorMappings "HosterCore/internal/app/rest_api_v2/pkg/error_mappings"
	JSONResponse "HosterCore/internal/app/rest_api_v2/pkg/json_response"
	HosterJailUtils "HosterCore/internal/pkg/hoster/jail/utils"
	HosterVmUtils "HosterCore/internal/pkg/hoster/vm/utils"
	zfsutils "HosterCore/internal/pkg/zfs_utils"
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
)

type SnapshotInput struct {
	SnapshotsToKeep int    `json:"snapshots_to_keep"`
	ResourceName    string `json:"res_name"`
	SnapshotType    string `json:"snapshot_type"`
	SnapshotDataset string `json:"-"`
}

// @Tags Snapshots
// @Summary Take a new snapshot.
// @Description Take a new VM or Jail snapshot, using the resource name (Jail name or a VM name).<br>`AUTH`: Only `rest` user is allowed.
// @Produce json
// @Security BasicAuth
// @Success 200 {object} SwaggerSuccess
// @Failure 500 {object} SwaggerError
// @Param Input body SnapshotInput true "Request payload"
// @Router /snapshot/take [post]
func SnapshotTake(w http.ResponseWriter, r *http.Request) {
	if !ApiAuth.CheckRestUser(r) {
		user, pass, _ := r.BasicAuth()
		UnauthenticatedResponse(w, user, pass)
		return
	}

	input := SnapshotInput{}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&input)
	if err != nil {
		ReportError(w, http.StatusInternalServerError, "could not parse your input")
		return
	}

	jails, err := HosterJailUtils.ListAllSimple()
	if err != nil {
		ReportError(w, http.StatusInternalServerError, err.Error())
		return
	}
	for _, v := range jails {
		if v.JailName == input.ResourceName {
			input.SnapshotDataset = v.DsName + "/" + v.JailName
		}
	}

	vms, err := HosterVmUtils.ListAllSimple()
	if err != nil {
		ReportError(w, http.StatusInternalServerError, err.Error())
		return
	}
	for _, v := range vms {
		if v.VmName == input.ResourceName {
			input.SnapshotDataset = v.DsName + "/" + v.VmName
		}
	}

	if len(input.SnapshotDataset) < 1 {
		ReportError(w, http.StatusInternalServerError, ErrorMappings.ResourceDoesntExist.String())
		return
	}

	// name, removed, err := zfsutils.TakeScheduledSnapshot(input.SnapshotDataset, input.SnapshotType, input.SnapshotsToKeep)
	_, _, err = zfsutils.TakeScheduledSnapshot(input.SnapshotDataset, input.SnapshotType, input.SnapshotsToKeep)
	if err != nil {
		ReportError(w, http.StatusInternalServerError, err.Error())
		return
	}

	payload, _ := JSONResponse.GenerateJson(w, "message", "success")
	SetStatusCode(w, http.StatusOK)
	w.Write(payload)
}

// @Tags Snapshots
// @Summary List all snapshots for any given VM or a Jail.
// @Description List all snapshots for any given VM or a Jail.<br>`AUTH`: Both users are allowed.
// @Produce json
// @Security BasicAuth
// @Success 200 {object} []zfsutils.SnapshotInfo
// @Failure 500 {object} SwaggerError
// @Param res_name path string true "Resource Name (Jail or VM)"
// @Router /snapshot/all/{res_name} [get]
func SnapshotList(w http.ResponseWriter, r *http.Request) {
	if !ApiAuth.CheckAnyUser(r) {
		user, pass, _ := r.BasicAuth()
		UnauthenticatedResponse(w, user, pass)
		return
	}

	vars := mux.Vars(r)
	resName := vars["res_name"]
	resDataset := ""

	jails, err := HosterJailUtils.ListAllSimple()
	if err != nil {
		ReportError(w, http.StatusInternalServerError, err.Error())
		return
	}
	vms, err := HosterVmUtils.ListAllSimple()
	if err != nil {
		ReportError(w, http.StatusInternalServerError, err.Error())
		return
	}

	for _, v := range jails {
		if v.JailName == resName {
			resDataset = v.DsName + "/" + v.JailName
		}
	}
	for _, v := range vms {
		if v.VmName == resName {
			resDataset = v.DsName + "/" + v.VmName
		}
	}
	if len(resDataset) < 1 {
		ReportError(w, http.StatusInternalServerError, ErrorMappings.ResourceDoesntExist.String())
		return
	}

	snaps, err := zfsutils.SnapshotListWithDescriptions()
	if err != nil {
		ReportError(w, http.StatusInternalServerError, err.Error())
		return
	}

	result := []zfsutils.SnapshotInfo{}
	for _, v := range snaps {
		if v.Dataset == resDataset {
			result = append(result, v)
		}
	}

	payload, err := json.Marshal(result)
	if err != nil {
		ReportError(w, http.StatusInternalServerError, err.Error())
		return
	}

	SetStatusCode(w, http.StatusOK)
	w.Write(payload)
}

// @Tags Snapshots
// @Summary Destroy a snapshot for any given VM or a Jail.
// @Description Destroy a snapshot for any given VM or a Jail.<br>`AUTH`: Only `rest` user is allowed.
// @Produce json
// @Security BasicAuth
// @Success 200 {object} SwaggerSuccess
// @Failure 500 {object} SwaggerError
// @Param res_name path string true "Snapshot Name"
// @Router /snapshot/destroy/{snapshot_name} [delete]
func SnapshotDestroy(w http.ResponseWriter, r *http.Request) {
	if !ApiAuth.CheckRestUser(r) {
		user, pass, _ := r.BasicAuth()
		UnauthenticatedResponse(w, user, pass)
		return
	}

	vars := mux.Vars(r)
	snapshotName := vars["snapshot_name"]

	snaps, err := zfsutils.SnapshotListWithDescriptions()
	if err != nil {
		ReportError(w, http.StatusInternalServerError, err.Error())
		return
	}

	snapFound := false
	for _, v := range snaps {
		if v.Name == snapshotName {
			err = zfsutils.RemoveSnapshot(snapshotName)
			if err != nil {
				ReportError(w, http.StatusInternalServerError, err.Error())
				return
			}
			snapFound = true
		}
	}
	if !snapFound {
		ReportError(w, http.StatusInternalServerError, ErrorMappings.SnapshotDoesntExist.String())
		return
	}

	payload, _ := JSONResponse.GenerateJson(w, "message", "success")
	SetStatusCode(w, http.StatusOK)
	w.Write(payload)
}
