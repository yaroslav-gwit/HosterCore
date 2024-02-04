package handlers

import (
	HosterJailUtils "HosterCore/internal/pkg/hoster/jail/utils"
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
)

// @Tags Jails
// @Summary List all Jails.
// @Description Get the list of all Jails, including the information about them.
// @Produce json
// @Success 200 {object} []HosterJailUtils.JailApi
// @Failure 500 {object} SwaggerError
// @Router /jail/all [get]
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

// @Tags Jails
// @Summary Get Jail info.
// @Description Get Jail info.
// @Produce json
// @Success 200 {object} HosterJailUtils.JailApi
// @Failure 500 {object} SwaggerError
// @Param jail_name path string true "Jail Name"
// @Router /jail/info/{jail_name} [get]
func JailInfo(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	jailName := vars["jail_name"]

	jails, err := HosterJailUtils.InfoJsonApi(jailName)
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
