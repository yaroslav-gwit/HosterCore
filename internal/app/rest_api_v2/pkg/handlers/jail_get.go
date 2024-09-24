package handlers

import (
	ApiAuth "HosterCore/internal/app/rest_api_v2/pkg/auth"
	HosterHost "HosterCore/internal/pkg/hoster/host"
	HosterJailUtils "HosterCore/internal/pkg/hoster/jail/utils"
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
)

// @Summary Get Jail config (settings).
// @Description Get Jail config (settings).<br>`AUTH`: Only `rest` user is allowed.
// @Produce json
// @Security BasicAuth
// @Success 200 {object} HosterJailUtils.JailConfig{}
// @Failure 500 {object} SwaggerError
// @Param jail_name path string true "Jail Name"
// @Router /jail/settings/{jail_name} [get]
func JailGetSettings(w http.ResponseWriter, r *http.Request) {
	if !ApiAuth.CheckRestUser(r) {
		user, pass, _ := r.BasicAuth()
		UnauthenticatedResponse(w, user, pass)
		return
	}

	vars := mux.Vars(r)
	jailName := vars["jail_name"]

	info, err := HosterJailUtils.InfoJsonApi(jailName)
	if err != nil {
		ReportError(w, http.StatusInternalServerError, err.Error())
		return
	}

	conf, err := HosterJailUtils.GetJailConfig(info.Simple.Mountpoint + "/" + info.Name)
	if err != nil {
		ReportError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// If DNS search domain is not set, get it from host config
	if len(conf.DnsSearchDomain) < 1 {
		hostConf, err := HosterHost.GetHostConfig()
		if err != nil {
			ReportError(w, http.StatusInternalServerError, err.Error())
			return
		}
		conf.DnsSearchDomain = hostConf.DnsSearchDomain
	}

	payload, err := json.Marshal(conf)
	if err != nil {
		ReportError(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	SetStatusCode(w, http.StatusOK)
	w.Write(payload)
}
