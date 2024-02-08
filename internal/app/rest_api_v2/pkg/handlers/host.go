package handlers

import (
	ApiAuth "HosterCore/internal/app/rest_api_v2/pkg/auth"
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
