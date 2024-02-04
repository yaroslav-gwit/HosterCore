package handlers

import (
	HosterJailUtils "HosterCore/internal/pkg/hoster/jail/utils"
	"encoding/json"
	"net/http"
)

// @Tags Jails
// @Summary List all Jails.
// @Summary Get the list of all Jails, including the information about them.
// @Produce json
// @Success 200 {object} []HosterJailUtils.JailApi
// @Failure 500 {object} SwaggerError
// @Router /vm/all [get]
func JailList(w http.ResponseWriter, r *http.Request) {
	jails, err := HosterJailUtils.ListJsonApi()
	if err != nil {
		ReportError(w, http.StatusInternalServerError, err.Error())
		return
	}

	payload, err := json.Marshal(jails)
	if err != nil {
		ReportError(w, http.StatusInternalServerError, err.Error())
		return
	}

	SetStatusCode(w, http.StatusOK)
	w.Write(payload)
}
