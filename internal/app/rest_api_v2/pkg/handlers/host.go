package handlers

import (
	ApiAuth "HosterCore/internal/app/rest_api_v2/pkg/auth"
	HosterHost "HosterCore/internal/pkg/hoster/host"
	HosterHostUtils "HosterCore/internal/pkg/hoster/host/utils"
	"encoding/json"
	"net/http"
)

// @Tags Host
// @Summary Get Host info.
// @Description Get Host info.<br>`AUTH`: Both users are allowed.
// @Produce json
// @Security BasicAuth
// @Success 200 {object} HosterHostUtils.HostInfo
// @Failure 500 {object} SwaggerError
// @Router /host/info [get]
func HostInfo(w http.ResponseWriter, r *http.Request) {
	if !ApiAuth.CheckAnyUser(r) {
		user, pass, _ := r.BasicAuth()
		UnauthenticatedResponse(w, user, pass)
		return
	}

	info, err := HosterHostUtils.GetHostInfo()
	if err != nil {
		ReportError(w, http.StatusInternalServerError, err.Error())
		return
	}

	payload, err := json.Marshal(info)
	if err != nil {
		ReportError(w, http.StatusInternalServerError, err.Error())
		return
	}

	SetStatusCode(w, http.StatusOK)
	w.Write(payload)
}

// @Tags Host
// @Summary Get Host Settings.
// @Description Get Host Settings.<br>`AUTH`: only REST user is allowed.
// @Produce json
// @Security BasicAuth
// @Success 200 {object} HosterHost.HostConfig
// @Failure 500 {object} SwaggerError
// @Router /host/settings [get]
func HostSettings(w http.ResponseWriter, r *http.Request) {
	if !ApiAuth.CheckRestUser(r) {
		user, pass, _ := r.BasicAuth()
		UnauthenticatedResponse(w, user, pass)
		return
	}

	info, err := HosterHost.GetHostConfig()
	if err != nil {
		ReportError(w, http.StatusInternalServerError, err.Error())
		return
	}

	payload, err := json.Marshal(info)
	if err != nil {
		ReportError(w, http.StatusInternalServerError, err.Error())
		return
	}

	SetStatusCode(w, http.StatusOK)
	w.Write(payload)
}
