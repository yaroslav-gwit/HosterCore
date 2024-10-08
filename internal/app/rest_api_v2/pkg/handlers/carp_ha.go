package handlers

import (
	CarpClient "HosterCore/internal/app/ha_carp/client"
	CarpUtils "HosterCore/internal/app/ha_carp/utils"
	ApiAuth "HosterCore/internal/app/rest_api_v2/pkg/auth"
	JSONResponse "HosterCore/internal/app/rest_api_v2/pkg/json_response"
	"encoding/json"
	"net/http"
)

// @Tags High Availability
// @Summary Ping CARP interface.
// @Description Ping CARP interface.<br>`AUTH`: Only HA user is allowed.
// @Produce json
// @Security BasicAuth
// @Success 200 {object} SwaggerSuccess
// @Failure 500 {object} SwaggerError
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

	payload, err := JSONResponse.GenerateJson(w, "message", "success")
	if err != nil {
		ReportError(w, http.StatusInternalServerError, err.Error())
		return
	}

	SetStatusCode(w, http.StatusOK)
	w.Write(payload)
}
