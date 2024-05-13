package handlers

import (
	ApiAuth "HosterCore/internal/app/rest_api_v2/pkg/auth"
	HosterNetwork "HosterCore/internal/pkg/hoster/network"
	"encoding/json"
	"net/http"
)

// @Tags Networks
// @Summary Get the networks list.
// @Description Get the networks list.<br>`AUTH`: Only REST user is allowed.
// @Produce json
// @Security BasicAuth
// @Success 200 {object} []HosterNetwork.NetworkConfig
// @Failure 500 {object} SwaggerError
// @Router /network/all [get]
func NetworkList(w http.ResponseWriter, r *http.Request) {
	if !ApiAuth.CheckRestUser(r) {
		user, pass, _ := r.BasicAuth()
		UnauthenticatedResponse(w, user, pass)
		return
	}

	info, err := HosterNetwork.GetNetworkConfig()
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
