// Copyright 2023 Hoster Authors. All rights reserved.
// Use of this source code is governed by an Apache License 2.0
// license that can be found in the LICENSE file.

package handlers

import (
	ApiAuth "HosterCore/internal/app/rest_api_v2/pkg/auth"
	ErrorMappings "HosterCore/internal/app/rest_api_v2/pkg/error_mappings"
	JSONResponse "HosterCore/internal/app/rest_api_v2/pkg/json_response"
	SchedulerClient "HosterCore/internal/app/scheduler/client"
	HosterJailUtils "HosterCore/internal/pkg/hoster/jail/utils"
	HosterVmUtils "HosterCore/internal/pkg/hoster/vm/utils"
	zfsutils "HosterCore/internal/pkg/zfs_utils"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/mux"
)

type SnapshotInput struct {
	SnapshotsToKeep int    `json:"snapshots_to_keep"` // How many snapshots to keep, e.g. 5
	SnapshotName    string `json:"snapshot_name"`     // Full snapshot name, including the whole path, e.g. "tank/vm-encrypted/vmTest1@snap1"
	ResourceName    string `json:"res_name"`          // VM or Jail name
	NewResourceName string `json:"new_res_name"`      // Used in clone operation, e.g. newVmName, the internal call will automatically append the dataset name
	SnapshotType    string `json:"snapshot_type"`     // "hourly", "daily", "weekly", "monthly", "frequent"
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
		ReportError(w, http.StatusInternalServerError, ErrorMappings.CouldNotParseYourInput.String())
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
// @Summary List all snapshots for any given VM or a Jail (cached version).
// @Description List all snapshots for any given VM or a Jail (cached version).<br>`AUTH`: Both users are allowed.
// @Produce json
// @Security BasicAuth
// @Success 200 {object} []zfsutils.SnapshotInfo
// @Failure 500 {object} SwaggerError
// @Param res_name path string true "Resource Name (Jail or VM)"
// @Router /snapshot/all/{res_name}/cache [get]
func SnapshotListCache(w http.ResponseWriter, r *http.Request) {
	if !ApiAuth.CheckAnyUser(r) {
		user, pass, _ := r.BasicAuth()
		UnauthenticatedResponse(w, user, pass)
		return
	}

	vars := mux.Vars(r)
	resName := vars["res_name"]
	resDataset := ""

	jails, err := HosterJailUtils.ReadCache()
	if err != nil {
		ReportError(w, http.StatusInternalServerError, err.Error())
		return
	}
	vms, err := HosterVmUtils.ReadCache()
	if err != nil {
		ReportError(w, http.StatusInternalServerError, err.Error())
		return
	}

	for _, v := range jails {
		if v.Name == resName {
			resDataset = v.Simple.DsName + "/" + v.Name
		}
	}
	if len(resDataset) < 1 {
		for _, v := range vms {
			if v.Name == resName {
				resDataset = v.Simple.DsName + "/" + v.Name
			}
		}
	}
	if len(resDataset) < 1 {
		ReportError(w, http.StatusInternalServerError, ErrorMappings.ResourceDoesntExist.String())
		return
	}

	snaps, err := zfsutils.ReadSnapshotCache()
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

	// Debugging
	// _ = result
	// mp := map[string]string{"ds": resDataset}
	// payload, err := json.Marshal(mp)

	payload, err := json.Marshal(result)
	if err != nil {
		ReportError(w, http.StatusInternalServerError, err.Error())
		return
	}

	SetStatusCode(w, http.StatusOK)
	w.Write(payload)
}

type SnapshotName struct {
	ResourceName string `json:"resource_name"` // VM or Jail name
	SnapshotName string `json:"snapshot_name"` // Full snapshot name, including the whole path, e.g. "tank/vm-encrypted/vmTest1@snap1"
}

// @Tags Snapshots
// @Summary Destroy a snapshot for any given VM or a Jail.
// @Description Destroy a snapshot for any given VM or a Jail.<br>`AUTH`: Only `rest` user is allowed.
// @Produce json
// @Security BasicAuth
// @Success 200 {object} SwaggerSuccess
// @Failure 500 {object} SwaggerError
// @Param Input body SnapshotName true "Request payload"
// @Router /snapshot/destroy [delete]
func SnapshotDestroy(w http.ResponseWriter, r *http.Request) {
	if !ApiAuth.CheckRestUser(r) {
		user, pass, _ := r.BasicAuth()
		UnauthenticatedResponse(w, user, pass)
		return
	}

	input := SnapshotName{}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&input)
	if err != nil {
		ReportError(w, http.StatusInternalServerError, ErrorMappings.CouldNotParseYourInput.String())
		return
	}

	jobID, err := SchedulerClient.AddSnapshotDestroyJob(input.ResourceName, input.SnapshotName)
	if err != nil {
		ReportError(w, http.StatusInternalServerError, err.Error())
		return
	}

	iterations := 0
	for {
		if iterations > 8 {
			ReportError(w, http.StatusInternalServerError, "job is still running in the background, but it's taking too long, please check the status manually")
			return
		}
		iterations++

		time.Sleep(1 * time.Second)

		jobStatus, err := SchedulerClient.GetJobInfo(jobID)
		if err != nil {
			ReportError(w, http.StatusInternalServerError, err.Error())
			return
		}

		if jobStatus.JobDone {
			payload, _ := JSONResponse.GenerateJson(w, "message", "success")
			SetStatusCode(w, http.StatusOK)
			w.Write(payload)
			return
		} else if jobStatus.JobFailed {
			ReportError(w, http.StatusInternalServerError, jobStatus.JobError)
			return
		} else {
			continue
		}
	}
}

// @Tags Snapshots
// @Summary Rollback to a previous snapshot.
// @Description Rollback to a previous snapshot.<br>`AUTH`: Only `rest` user is allowed.<br>
// @Produce json
// @Security BasicAuth
// @Success 200 {object} SwaggerSuccess
// @Failure 500 {object} SwaggerError
// @Param Input body SnapshotName true "Request payload"
// @Router /snapshot/rollback [post]
func SnapshotRollback(w http.ResponseWriter, r *http.Request) {
	if !ApiAuth.CheckRestUser(r) {
		user, pass, _ := r.BasicAuth()
		UnauthenticatedResponse(w, user, pass)
		return
	}

	input := SnapshotName{}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&input)
	if err != nil {
		ReportError(w, http.StatusInternalServerError, ErrorMappings.CouldNotParseYourInput.String())
		return
	}

	jobID, err := SchedulerClient.AddSnapshotRollbackJob(input.ResourceName, input.SnapshotName)
	if err != nil {
		ReportError(w, http.StatusInternalServerError, err.Error())
		return
	}

	iterations := 0
	for {
		if iterations > 8 {
			ReportError(w, http.StatusInternalServerError, "job is still running in the background, but it's taking too long, please check the status manually")
			return
		}
		iterations++

		time.Sleep(1 * time.Second)

		jobStatus, err := SchedulerClient.GetJobInfo(jobID)
		if err != nil {
			ReportError(w, http.StatusInternalServerError, err.Error())
			return
		}

		if jobStatus.JobDone {
			payload, _ := JSONResponse.GenerateJson(w, "message", "success")
			SetStatusCode(w, http.StatusOK)
			w.Write(payload)
			return
		} else if jobStatus.JobFailed {
			ReportError(w, http.StatusInternalServerError, jobStatus.JobError)
			return
		} else {
			continue
		}
	}
}

// @Tags Snapshots
// @Summary Clone an existing VM or Jail snapshot.
// @Description Clone an existing VM or Jail snapshot.<br>`AUTH`: Only `rest` user is allowed.<br>
// @Produce json
// @Security BasicAuth
// @Success 200 {object} SwaggerSuccess
// @Failure 500 {object} SwaggerError
// @Param Input body SnapshotInput true "Request payload"
// @Router /snapshot/clone [post]
func SnapshotClone(w http.ResponseWriter, r *http.Request) {
	if !ApiAuth.CheckRestUser(r) {
		user, pass, _ := r.BasicAuth()
		UnauthenticatedResponse(w, user, pass)
		return
	}

	input := SnapshotInput{}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&input)
	if err != nil {
		ReportError(w, http.StatusInternalServerError, ErrorMappings.CouldNotParseYourInput.String())
		return
	}

	tempLs := strings.Split(input.SnapshotName, "/")
	if len(tempLs) < 2 {
		ReportError(w, http.StatusInternalServerError, ErrorMappings.CouldNotParseYourInput.String())
		return
	}

	newRes := strings.Join(tempLs[:len(tempLs)-2], "/") + "/" + input.NewResourceName
	err = zfsutils.SnapshotClone(input.SnapshotName, newRes)
	if err != nil {
		ReportError(w, http.StatusInternalServerError, ErrorMappings.CouldNotParseYourInput.String())
		return
	}

	payload, _ := JSONResponse.GenerateJson(w, "message", "success")
	SetStatusCode(w, http.StatusOK)
	w.Write(payload)
}
