package handlers

import (
	CarpClient "HosterCore/internal/app/ha_carp/client"
	CarpUtils "HosterCore/internal/app/ha_carp/utils"
	ApiAuth "HosterCore/internal/app/rest_api_v2/pkg/auth"
	FreeBSDsysctls "HosterCore/internal/pkg/freebsd/sysctls"
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

	err = CarpClient.HostAdd(input)
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

	// Make and populate the response map
	mp := make(map[string]interface{})
	mp["message"] = "success"
	payload, err := json.Marshal(mp)
	if err != nil {
		ReportError(w, http.StatusInternalServerError, err.Error())
		return
	}

	SetStatusCode(w, http.StatusOK)
	w.Write(payload)
}
