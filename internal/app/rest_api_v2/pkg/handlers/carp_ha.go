package handlers

import (
	CarpClient "HosterCore/internal/app/ha_carp/client"
	CarpUtils "HosterCore/internal/app/ha_carp/utils"
	ApiAuth "HosterCore/internal/app/rest_api_v2/pkg/auth"
	JSONResponse "HosterCore/internal/app/rest_api_v2/pkg/json_response"
	FreeBSDsysctls "HosterCore/internal/pkg/freebsd/sysctls"
	HosterJailUtils "HosterCore/internal/pkg/hoster/jail/utils"
	HosterVmUtils "HosterCore/internal/pkg/hoster/vm/utils"
	"encoding/json"
	"net/http"
	"strings"
)

// @Tags High Availability
// @Summary Ping CARP interface.
// @Description Ping CARP interface.<br>`AUTH`: Only HA user is allowed.
// @Produce json
// @Security BasicAuth
// @Success 200 {object} CarpUtils.CarpPingResponse{}
// @Failure 500 {object} SwaggerError{}
// @Param Input body CarpUtils.HostInfo{} true "Request Payload"
// @Router /carp-ha/ping [post]
func CarpPing(w http.ResponseWriter, r *http.Request) {
	if !ApiAuth.CheckHaUser(r) {
		user, pass, _ := r.BasicAuth()
		UnauthenticatedResponse(w, user, pass)
		return
	}

	input := CarpUtils.HostInfo{}
	input.IpAddress = r.RemoteAddr
	input.IpAddress = strings.Split(r.RemoteAddr, ":")[0]

	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&input)
	if err != nil {
		ReportError(w, http.StatusInternalServerError, err.Error())
		return
	}

	err = CarpClient.ReceiveHostAdd(input)
	if err != nil {
		ReportError(w, http.StatusInternalServerError, err.Error())
		return
	}

	hostname, err := FreeBSDsysctls.SysctlKernHostname()
	if err != nil {
		ReportError(w, http.StatusInternalServerError, err.Error())
		return
	}

	res := CarpUtils.CarpPingResponse{
		Message:  "success",
		Hostname: hostname,
	}

	payload, err := json.Marshal(res)
	if err != nil {
		ReportError(w, http.StatusInternalServerError, err.Error())
		return
	}

	SetStatusCode(w, http.StatusOK)
	w.Write(payload)
}

// @Tags High Availability
// @Summary Receive the cluster state from the master.
// @Description Receive the cluster state from the master.<br>`AUTH`: Only HA user is allowed.
// @Produce json
// @Security BasicAuth
// @Success 200 {object} SwaggerSuccess{}
// @Failure 500 {object} SwaggerError{}
// @Param Input body CarpUtils.HaStatus{} true "Request Payload"
// @Param master_hostname path string true "Hostname of the master server"
// @Router /carp-ha/receive-state/{master_hostname} [post]
func CarpReceiveHostState(w http.ResponseWriter, r *http.Request) {
	if !ApiAuth.CheckHaUser(r) {
		user, pass, _ := r.BasicAuth()
		UnauthenticatedResponse(w, user, pass)
		return
	}

	input := CarpUtils.HaStatus{}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&input)
	if err != nil {
		ReportError(w, http.StatusInternalServerError, err.Error())
		return
	}

	err = CarpClient.ReceiveRemoteState(input)
	if err != nil {
		ReportError(w, http.StatusInternalServerError, err.Error())
		return
	}

	payload, _ := JSONResponse.GenerateJson(w, "message", "success")
	SetStatusCode(w, http.StatusOK)
	w.Write(payload)
}

// @Tags High Availability
// @Summary Receive the cluster state from the master.
// @Description Receive the cluster state from the master.<br>`AUTH`: Only HA user is allowed.
// @Produce json
// @Security BasicAuth
// @Success 200 {object} []CarpUtils.BackupInfo{}
// @Failure 500 {object} SwaggerError{}
// @Router /carp-ha/backups [get]
func CarpReturnListOfBackups(w http.ResponseWriter, r *http.Request) {
	if !ApiAuth.CheckHaUser(r) {
		user, pass, _ := r.BasicAuth()
		UnauthenticatedResponse(w, user, pass)
		return
	}

	haConf, err := CarpUtils.ParseCarpConfigFile()
	if err != nil {
		ReportError(w, http.StatusInternalServerError, err.Error())
		return
	}

	backups := []CarpUtils.BackupInfo{}
	if haConf.ParticipateInFailover {
		vms, err := HosterVmUtils.ReadCache()
		if err != nil {
			ReportError(w, http.StatusInternalServerError, err.Error())
			return
		}
		jails, err := HosterJailUtils.ReadCache()
		if err != nil {
			ReportError(w, http.StatusInternalServerError, err.Error())
			return
		}

		for _, v := range vms {
			if v.Backup {
				temp := CarpUtils.BackupInfo{}
				temp.CurrentHost = v.CurrentHost
				temp.ResourceType = "vm"
				temp.ResourceName = v.Name
				temp.ParentHost = v.ParentHost

				backups = append(backups, temp)
			}
		}

		for _, v := range jails {
			if v.Backup {
				temp := CarpUtils.BackupInfo{}
				temp.CurrentHost = v.CurrentHost
				temp.ResourceType = "jail"
				temp.ResourceName = v.Name
				temp.ParentHost = v.Parent

				backups = append(backups, temp)
			}
		}
	}

	// snaps, err := zfsutils.ReadSnapshotCache()
	// if err != nil {
	// 	ReportError(w, http.StatusInternalServerError, err.Error())
	// 	return
	// }

	// result := []zfsutils.SnapshotInfo{}
	// for _, v := range snaps {
	// 	if v.Dataset == resDataset {
	// 		result = append(result, v)
	// 	}
	// }

	payload, err := json.Marshal(backups)
	if err != nil {
		ReportError(w, http.StatusInternalServerError, err.Error())
		return
	}

	SetStatusCode(w, http.StatusOK)
	w.Write(payload)
}
