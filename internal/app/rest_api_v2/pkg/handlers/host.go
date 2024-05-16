package handlers

import (
	ApiAuth "HosterCore/internal/app/rest_api_v2/pkg/auth"
	RestApiConfig "HosterCore/internal/app/rest_api_v2/pkg/config"
	JSONResponse "HosterCore/internal/app/rest_api_v2/pkg/json_response"
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

// @Tags Host
// @Summary Get RestAPI Settings (including HA settings).
// @Description Get RestAPI Settings (including HA settings).<br>`AUTH`: only REST user is allowed.
// @Produce json
// @Security BasicAuth
// @Success 200 {object} RestApiConfig.RestApiConfig
// @Failure 500 {object} SwaggerError
// @Router /host/settings/api [get]
func HostRestApiSettings(w http.ResponseWriter, r *http.Request) {
	if !ApiAuth.CheckRestUser(r) {
		user, pass, _ := r.BasicAuth()
		UnauthenticatedResponse(w, user, pass)
		return
	}

	info, err := RestApiConfig.GetApiConfig()
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

type DnsSearchDomainInput struct {
	DnsSearchDomain string `json:"dns_search_domain"`
}

// @Tags Host
// @Summary Post a new DNS search domain.
// @Description Post a new DNS search domain.
// @Produce json
// @Security BasicAuth
// @Success 200 {object} SwaggerSuccess
// @Failure 500 {object} SwaggerError
// @Param Input body DnsSearchDomainInput true "Request Payload"
// @Router /host/settings/dns-search-domain [post]
func PostHostSettingsDnsSearchDomain(w http.ResponseWriter, r *http.Request) {
	if !ApiAuth.CheckRestUser(r) {
		user, pass, _ := r.BasicAuth()
		UnauthenticatedResponse(w, user, pass)
		return
	}

	input := DnsSearchDomainInput{}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&input)
	if err != nil {
		ReportError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if len(input.DnsSearchDomain) < 1 {
		ReportError(w, http.StatusBadRequest, "dns_search_domain is required")
		return
	}

	hostConf, err := HosterHost.GetHostConfig()
	if err != nil {
		ReportError(w, http.StatusInternalServerError, err.Error())
		return
	}

	hostConf.DnsSearchDomain = input.DnsSearchDomain
	err = HosterHost.SaveHostConfig(hostConf)
	if err != nil {
		ReportError(w, http.StatusInternalServerError, err.Error())
		return
	}

	payload, _ := JSONResponse.GenerateJson(w, "message", "success")
	SetStatusCode(w, http.StatusOK)
	w.Write(payload)
}

type VmTemplateLink struct {
	Link string `json:"vm_template_link"`
}

// @Tags Host
// @Summary Post an updated VM template site.
// @Description Post an updated VM template site.
// @Produce json
// @Security BasicAuth
// @Success 200 {object} SwaggerSuccess
// @Failure 500 {object} SwaggerError
// @Param Input body VmTemplateLink true "Request Payload"
// @Router /host/settings/vm-templates [post]
func PostHostSettingsVmTemplateLink(w http.ResponseWriter, r *http.Request) {
	if !ApiAuth.CheckRestUser(r) {
		user, pass, _ := r.BasicAuth()
		UnauthenticatedResponse(w, user, pass)
		return
	}

	input := VmTemplateLink{}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&input)
	if err != nil {
		ReportError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if len(input.Link) < 1 {
		ReportError(w, http.StatusBadRequest, "vm_template_link is required")
		return
	}

	hostConf, err := HosterHost.GetHostConfig()
	if err != nil {
		ReportError(w, http.StatusInternalServerError, err.Error())
		return
	}

	hostConf.ImageServer = input.Link
	err = HosterHost.SaveHostConfig(hostConf)
	if err != nil {
		ReportError(w, http.StatusInternalServerError, err.Error())
		return
	}

	payload, _ := JSONResponse.GenerateJson(w, "message", "success")
	SetStatusCode(w, http.StatusOK)
	w.Write(payload)
}
