package handlers

import (
	ApiAuth "HosterCore/internal/app/rest_api_v2/pkg/auth"
	RestApiConfig "HosterCore/internal/app/rest_api_v2/pkg/config"
	JSONResponse "HosterCore/internal/app/rest_api_v2/pkg/json_response"
	HosterHost "HosterCore/internal/pkg/hoster/host"
	HosterHostUtils "HosterCore/internal/pkg/hoster/host/utils"
	"encoding/json"
	"net/http"
	"os"
	"slices"
	"strings"

	"github.com/gorilla/mux"
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

type UpstreamDnsInput struct {
	DnsServer string `json:"dns_server"`
}

// @Tags Host
// @Summary Add a new upstream DNS server.
// @Description Add a new upstream DNS server.
// @Produce json
// @Security BasicAuth
// @Success 200 {object} SwaggerSuccess
// @Failure 500 {object} SwaggerError
// @Param Input body UpstreamDnsInput true "Request Payload"
// @Router /host/settings/add-upstream-dns [post]
func PostHostSettingsAddUpstreamDns(w http.ResponseWriter, r *http.Request) {
	if !ApiAuth.CheckRestUser(r) {
		user, pass, _ := r.BasicAuth()
		UnauthenticatedResponse(w, user, pass)
		return
	}

	input := UpstreamDnsInput{}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&input)
	if err != nil {
		ReportError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if len(input.DnsServer) < 1 {
		ReportError(w, http.StatusBadRequest, "dns_server is required")
		return
	}

	hostConf, err := HosterHost.GetHostConfig()
	if err != nil {
		ReportError(w, http.StatusInternalServerError, err.Error())
		return
	}

	hostConf.DnsServers = append(hostConf.DnsServers, input.DnsServer)
	err = HosterHost.SaveHostConfig(hostConf)
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

// @Tags Host
// @Summary Delete an upstream DNS server.
// @Description Delete an upstream DNS server.
// @Produce json
// @Security BasicAuth
// @Success 200 {object} SwaggerSuccess
// @Failure 500 {object} SwaggerError
// @Param Input body UpstreamDnsInput true "Request Payload"
// @Router /host/settings/delete-upstream-dns [delete]
func DeleteHostSettingsUpstreamDns(w http.ResponseWriter, r *http.Request) {
	if !ApiAuth.CheckRestUser(r) {
		user, pass, _ := r.BasicAuth()
		UnauthenticatedResponse(w, user, pass)
		return
	}

	input := UpstreamDnsInput{}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&input)
	if err != nil {
		ReportError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if len(input.DnsServer) < 1 {
		ReportError(w, http.StatusBadRequest, "dns_server is required")
		return
	}

	hostConf, err := HosterHost.GetHostConfig()
	if err != nil {
		ReportError(w, http.StatusInternalServerError, err.Error())
		return
	}

	servers := []string{}
	for _, v := range hostConf.DnsServers {
		if v != input.DnsServer {
			servers = append(servers, v)
		}
	}
	hostConf.DnsServers = servers

	err = HosterHost.SaveHostConfig(hostConf)
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

type SshKeyInput struct {
	KeyComment string `json:"key_comment"`
	KeyValue   string `json:"key_value"`
}

// @Tags Host
// @Summary Add a new VM SSH access key.
// @Description Add a new VM SSH access key.
// @Produce json
// @Security BasicAuth
// @Success 200 {object} SwaggerSuccess
// @Failure 500 {object} SwaggerError
// @Param Input body SshKeyInput true "Request Payload"
// @Router /host/settings/add-ssh-key [post]
func PostHostSettingsSshKey(w http.ResponseWriter, r *http.Request) {
	if !ApiAuth.CheckRestUser(r) {
		user, pass, _ := r.BasicAuth()
		UnauthenticatedResponse(w, user, pass)
		return
	}

	input := SshKeyInput{}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&input)
	if err != nil {
		ReportError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if len(input.KeyComment) < 1 {
		ReportError(w, http.StatusBadRequest, "key_comment is required")
		return
	}
	if len(input.KeyValue) < 1 {
		ReportError(w, http.StatusBadRequest, "key_value is required")
		return
	}

	hostConf, err := HosterHost.GetHostConfig()
	if err != nil {
		ReportError(w, http.StatusInternalServerError, err.Error())
		return
	}

	hostConf.HostSSHKeys = append(hostConf.HostSSHKeys, HosterHost.HostConfigKey{KeyValue: input.KeyValue, Comment: input.KeyComment})
	err = HosterHost.SaveHostConfig(hostConf)
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

// @Tags Host
// @Summary Delete an existing SSH key.
// @Description Delete an existing SSH key.
// @Produce json
// @Security BasicAuth
// @Success 200 {object} SwaggerSuccess
// @Failure 500 {object} SwaggerError
// @Param Input body SshKeyInput true "Request Payload"
// @Router /host/settings/delete-ssh-key [delete]
func DeleteHostSettingsSshKey(w http.ResponseWriter, r *http.Request) {
	if !ApiAuth.CheckRestUser(r) {
		user, pass, _ := r.BasicAuth()
		UnauthenticatedResponse(w, user, pass)
		return
	}

	input := SshKeyInput{}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&input)
	if err != nil {
		ReportError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if len(input.KeyValue) < 1 {
		ReportError(w, http.StatusBadRequest, "key_value is required")
		return
	}

	hostConf, err := HosterHost.GetHostConfig()
	if err != nil {
		ReportError(w, http.StatusInternalServerError, err.Error())
		return
	}

	hostKeys := []HosterHost.HostConfigKey{}
	for _, v := range hostConf.HostSSHKeys {
		if v.KeyValue != input.KeyValue {
			hostKeys = append(hostKeys, v)
		}
	}
	hostConf.HostSSHKeys = hostKeys

	err = HosterHost.SaveHostConfig(hostConf)
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

type HostAuthSshKeyInput struct {
	KeyValue string `json:"key_value"`
}

// @Tags Host
// @Summary Add a new host-level authorized SSH key.
// @Description Add a new host-level authorized SSH key.
// @Produce json
// @Security BasicAuth
// @Success 200 {object} SwaggerSuccess
// @Failure 500 {object} SwaggerError
// @Param Input body HostAuthSshKeyInput true "Request Payload"
// @Router /host/settings/ssh-auth-key [post]
func PostHostSshAuthKey(w http.ResponseWriter, r *http.Request) {
	authKeyLocation := "/root/.ssh/authorized_keys"
	if !ApiAuth.CheckAnyUser(r) {
		user, pass, _ := r.BasicAuth()
		UnauthenticatedResponse(w, user, pass)
		return
	}

	input := HostAuthSshKeyInput{}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&input)
	if err != nil {
		ReportError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if len(input.KeyValue) < 1 {
		ReportError(w, http.StatusBadRequest, "key_value is required")
		return
	}

	keyFile, err := os.ReadFile(authKeyLocation)
	if err != nil {
		ReportError(w, http.StatusInternalServerError, err.Error())
		return
	}

	payload, err := JSONResponse.GenerateJson(w, "message", "success")
	if err != nil {
		ReportError(w, http.StatusInternalServerError, err.Error())
		return
	}

	kSplit := strings.Split(string(keyFile), "\n")
	keys := []string{}
	for _, v := range kSplit {
		if len(strings.TrimSpace(v)) > 0 {
			keys = append(keys, v)
		}
	}

	if slices.Contains(keys, input.KeyValue) {
		SetStatusCode(w, http.StatusOK)
		w.Write(payload)
	} else {
		keys = append(keys, input.KeyValue)
		err = os.WriteFile(authKeyLocation, []byte(strings.Join(keys, "\n")), 0600)
		if err != nil {
			ReportError(w, http.StatusInternalServerError, err.Error())
			return
		}

		SetStatusCode(w, http.StatusOK)
		w.Write(payload)
	}
}

// @Tags Tags
// @Summary Add a new Host-related tag.
// @Description Add a new Host-related tag.
// @Produce json
// @Security BasicAuth
// @Success 200 {object} SwaggerSuccess
// @Failure 500 {object} SwaggerError
// @Param tag path string true "Host Tag"
// @Router /host/settings/add-tag/{tag} [post]
func PostHostTag(w http.ResponseWriter, r *http.Request) {
	if !ApiAuth.CheckRestUser(r) {
		user, pass, _ := r.BasicAuth()
		UnauthenticatedResponse(w, user, pass)
		return
	}

	input := mux.Vars(r)["tag"]

	if len(input) < 1 {
		ReportError(w, http.StatusBadRequest, "tag is required")
		return
	}
	if len(input) > 20 {
		ReportError(w, http.StatusBadRequest, "tag length must be less than 20 characters")
		return
	}

	hostConf, err := HosterHost.GetHostConfig()
	if err != nil {
		ReportError(w, http.StatusInternalServerError, err.Error())
		return
	}

	hostConf.Tags = append(hostConf.Tags, input)
	err = HosterHost.SaveHostConfig(hostConf)
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

// @Tags Tags
// @Summary Delete an existing Host-related tag.
// @Description Delete an existing Host-related tag.
// @Produce json
// @Security BasicAuth
// @Success 200 {object} SwaggerSuccess
// @Failure 500 {object} SwaggerError
// @Param tag path string true "Host Tag"
// @Router /host/settings/delete-tag/{tag} [delete]
func DeleteHostTag(w http.ResponseWriter, r *http.Request) {
	if !ApiAuth.CheckRestUser(r) {
		user, pass, _ := r.BasicAuth()
		UnauthenticatedResponse(w, user, pass)
		return
	}

	input := mux.Vars(r)["tag"]

	if len(input) < 1 {
		ReportError(w, http.StatusBadRequest, "tag is required")
		return
	}
	if len(input) > 20 {
		ReportError(w, http.StatusBadRequest, "tag length must be less than 20 characters")
		return
	}

	hostConf, err := HosterHost.GetHostConfig()
	if err != nil {
		ReportError(w, http.StatusInternalServerError, err.Error())
		return
	}

	tags := []string{}
	for _, v := range hostConf.Tags {
		if v != input {
			tags = append(tags, v)
		}
	}
	hostConf.Tags = tags

	err = HosterHost.SaveHostConfig(hostConf)
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
