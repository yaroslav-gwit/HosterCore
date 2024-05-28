package handlers

import (
	ApiAuth "HosterCore/internal/app/rest_api_v2/pkg/auth"
	JSONResponse "HosterCore/internal/app/rest_api_v2/pkg/json_response"
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

// @Tags Networks
// @Summary Add a new network.
// @Description Add a new network.<br>`AUTH`: Only REST user is allowed.
// @Produce json
// @Security BasicAuth
// @Success 200 {object} SwaggerSuccess
// @Failure 500 {object} SwaggerError
// @Param Input body []HosterNetwork.NetworkConfig true "Request Payload"
// @Router /network/add-new-network [post]
func PostNewNetwork(w http.ResponseWriter, r *http.Request) {
	if !ApiAuth.CheckRestUser(r) {
		user, pass, _ := r.BasicAuth()
		UnauthenticatedResponse(w, user, pass)
		return
	}

	input := HosterNetwork.NetworkConfig{}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&input)
	if err != nil {
		ReportError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// if len(input.KeyComment) < 1 {
	// 	ReportError(w, http.StatusBadRequest, "key_comment is required")
	// 	return
	// }
	// if len(input.KeyValue) < 1 {
	// 	ReportError(w, http.StatusBadRequest, "key_value is required")
	// 	return
	// }

	netConf, err := HosterNetwork.GetNetworkConfig()
	if err != nil {
		ReportError(w, http.StatusInternalServerError, err.Error())
		return
	}

	netConf = append(netConf, input)
	err = HosterNetwork.SaveNetworkConfig(netConf)
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
