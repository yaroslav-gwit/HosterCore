package handlers

import (
	ApiAuth "HosterCore/internal/app/rest_api_v2/pkg/auth"
	JSONResponse "HosterCore/internal/app/rest_api_v2/pkg/json_response"
	HosterJailUtils "HosterCore/internal/pkg/hoster/jail/utils"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
)

// @Tags Jails
// @Summary Update Jails's description.
// @Description Update Jails's description.<br>`AUTH`: Only `rest` user is allowed.
// @Produce json
// @Security BasicAuth
// @Success 200 {object} SwaggerSuccess
// @Failure 500 {object} SwaggerError
// @Param jail_name path string true "Name of the Jail"
// @Param Input body ResourceDescription{} true "Request payload"
// @Router /jail/settings/description/{jail_name} [post]
func JailPostDescription(w http.ResponseWriter, r *http.Request) {
	if !ApiAuth.CheckRestUser(r) {
		user, pass, _ := r.BasicAuth()
		UnauthenticatedResponse(w, user, pass)
		return
	}

	vars := mux.Vars(r)
	jailName := vars["jail_name"]

	input := ResourceDescription{}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&input)
	if err != nil {
		ReportError(w, http.StatusInternalServerError, err.Error())
		return
	}

	err = HosterJailUtils.UpdateDescription(jailName, input.Description)
	if err != nil {
		ReportError(w, http.StatusInternalServerError, err.Error())
		return
	}

	payload, _ := JSONResponse.GenerateJson(w, "message", "success")
	SetStatusCode(w, http.StatusOK)
	w.Write(payload)
}

// @Tags Jails, Tags
// @Summary Add a new tag for any particular Jail.
// @Description Add a new tag for any particular Jail.<br>`AUTH`: Only `rest` user is allowed.
// @Produce json
// @Security BasicAuth
// @Success 200 {object} SwaggerSuccess
// @Failure 500 {object} SwaggerError
// @Param jail_name path string true "Jail Name"
// @Param new_tag path string true "New Tag"
// @Param Input body TagInput true "Request payload"
// @Router /jail/settings/add-tag/{jail_name} [post]
func JailPostNewTag(w http.ResponseWriter, r *http.Request) {
	if !ApiAuth.CheckRestUser(r) {
		user, pass, _ := r.BasicAuth()
		UnauthenticatedResponse(w, user, pass)
		return
	}

	vars := mux.Vars(r)
	jailName := vars["jail_name"]

	input := TagInput{}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&input)
	if err != nil {
		ReportError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if len(input.Tag) < 1 {
		ReportError(w, http.StatusBadRequest, "tag must be at least 1 character long")
		return
	}

	if len(input.Tag) > 255 {
		ReportError(w, http.StatusBadRequest, "tag must be at most 255 characters long")
		return
	}

	jailInfo, err := HosterJailUtils.InfoJsonApi(jailName)
	if err != nil {
		ReportError(w, http.StatusInternalServerError, err.Error())
		return
	}
	location := jailInfo.Simple.Mountpoint + "/" + jailInfo.Name + "/" + HosterJailUtils.JAIL_CONFIG_NAME

	var tagFound bool
	for _, v := range jailInfo.JailConfig.Tags {
		if v == input.Tag {
			tagFound = true
			break
		}
	}
	if !tagFound {
		jailInfo.JailConfig.Tags = append(jailInfo.JailConfig.Tags, input.Tag)
	}

	err = HosterJailUtils.ConfigFileWriter(jailInfo.JailConfig, location)
	if err != nil {
		ReportError(w, http.StatusInternalServerError, err.Error())
		return
	}

	payload, _ := JSONResponse.GenerateJson(w, "message", "success")
	SetStatusCode(w, http.StatusOK)
	w.Write(payload)

	SetStatusCode(w, http.StatusOK)
	w.Write(payload)
}

// @Tags Jails
// @Summary Modify Jail's Workload type (e.g. is this a production Jail, true or false).
// @Description Modify Jail's Workload type (e.g. is this a production Jail, true or false).<br>`AUTH`: Only `rest` user is allowed.
// @Produce json
// @Security BasicAuth
// @Success 200 {object} SwaggerSuccess
// @Failure 500 {object} SwaggerError
// @Param jail_name path string true "Jail Name"
// @Param production path string true "Workload type (is this a production Jail?), e.g. true or false"
// @Router /jail/settings/production/{jail_name}/{production} [post]
func JailPostProductionSetting(w http.ResponseWriter, r *http.Request) {
	if !ApiAuth.CheckRestUser(r) {
		user, pass, _ := r.BasicAuth()
		UnauthenticatedResponse(w, user, pass)
		return
	}

	vars := mux.Vars(r)
	jailName := vars["jail_name"]
	prod := vars["production"]

	prod = strings.ToLower(prod)
	if prod == "true" || prod == "false" {
		_ = 0
	} else {
		errValue := "invalid workload type, must be either 'true' or 'false'"
		ReportError(w, http.StatusInternalServerError, errValue)
		return
	}

	jailInfo, err := HosterJailUtils.InfoJsonApi(jailName)
	if err != nil {
		ReportError(w, http.StatusInternalServerError, err.Error())
		return
	}
	location := jailInfo.Simple.Mountpoint + "/" + jailName + "/" + HosterJailUtils.JAIL_CONFIG_NAME

	if prod == "true" {
		jailInfo.JailConfig.Production = true
	} else {
		jailInfo.JailConfig.Production = false
	}
	err = HosterJailUtils.ConfigFileWriter(jailInfo.JailConfig, location)
	if err != nil {
		ReportError(w, http.StatusInternalServerError, err.Error())
		return
	}

	payload, _ := JSONResponse.GenerateJson(w, "message", "success")
	SetStatusCode(w, http.StatusOK)
	w.Write(payload)
}

// @Tags Jails
// @Summary Modify Jail's CPU limitation (in %, 1-100).
// @Description Modify Jail's CPU limitation (in %, maximum 100).<br>`AUTH`: Only `rest` user is allowed.
// @Produce json
// @Security BasicAuth
// @Success 200 {object} SwaggerSuccess
// @Failure 500 {object} SwaggerError
// @Param jail_name path string true "Jail Name"
// @Param limit path int true "Percentage limit (1-100)"
// @Router /jail/settings/cpu/{jail_name}/{limit} [post]
func JailPostCpuPercentageLimit(w http.ResponseWriter, r *http.Request) {
	if !ApiAuth.CheckRestUser(r) {
		user, pass, _ := r.BasicAuth()
		UnauthenticatedResponse(w, user, pass)
		return
	}

	vars := mux.Vars(r)
	jailName := vars["jail_name"]
	limit := vars["limit"]

	limitInt, err := strconv.Atoi(limit)
	if err != nil {
		ReportError(w, http.StatusInternalServerError, "limit value is not an integer: "+err.Error())
		return
	}

	if limitInt < 1 || limitInt > 100 {
		errValue := "invalid CPU limit, value must be between 1 and 100"
		ReportError(w, http.StatusInternalServerError, errValue)
		return
	}

	jailInfo, err := HosterJailUtils.InfoJsonApi(jailName)
	if err != nil {
		ReportError(w, http.StatusInternalServerError, err.Error())
		return
	}

	jailInfo.JailConfig.CPULimitPercent = limitInt
	location := jailInfo.Simple.Mountpoint + "/" + jailName + "/" + HosterJailUtils.JAIL_CONFIG_NAME
	err = HosterJailUtils.ConfigFileWriter(jailInfo.JailConfig, location)
	if err != nil {
		ReportError(w, http.StatusInternalServerError, err.Error())
		return
	}

	payload, _ := JSONResponse.GenerateJson(w, "message", "success")
	SetStatusCode(w, http.StatusOK)
	w.Write(payload)
}

// @Tags Jails
// @Summary Modify Jail's RAM limit.
// @Description Modify Jail's RAM limit.<br>`AUTH`: Only `rest` user is allowed.
// @Produce json
// @Security BasicAuth
// @Success 200 {object} SwaggerSuccess
// @Failure 500 {object} SwaggerError
// @Param jail_name path string true "Jail Name"
// @Param limit path string true "Memory limit (in MB or GB, e.g. 2GB, or 2048MB)"
// @Router /jail/settings/ram/{jail_name}/{limit} [post]
func JailPostRamLimit(w http.ResponseWriter, r *http.Request) {
	if !ApiAuth.CheckRestUser(r) {
		user, pass, _ := r.BasicAuth()
		UnauthenticatedResponse(w, user, pass)
		return
	}

	vars := mux.Vars(r)
	jailName := vars["jail_name"]
	limit := vars["limit"]

	if !strings.HasSuffix(limit, "M") || !strings.HasSuffix(limit, "MB") {
		if !strings.HasSuffix(limit, "G") || !strings.HasSuffix(limit, "GB") {
			errValue := "invalid RAM limit, must end with 'M', 'MB', 'G', or 'GB'"
			ReportError(w, http.StatusInternalServerError, errValue)
			return
		}
	}

	limitType := ""
	if strings.HasSuffix(limit, "M") || strings.HasSuffix(limit, "MB") {
		limit = strings.TrimSuffix(limit, "MB")
		limit = strings.TrimSuffix(limit, "M")
		limitType = "M"
	} else {
		limit = strings.TrimSuffix(limit, "GB")
		limit = strings.TrimSuffix(limit, "G")
		limitType = "G"
	}

	limitInt, err := strconv.Atoi(limit)
	if err != nil {
		ReportError(w, http.StatusInternalServerError, "memory limit value could not be parsed: "+err.Error())
		return
	}

	jailInfo, err := HosterJailUtils.InfoJsonApi(jailName)
	if err != nil {
		ReportError(w, http.StatusInternalServerError, err.Error())
		return
	}

	jailInfo.JailConfig.RAMLimit = fmt.Sprintf("%d%s", limitInt, limitType)
	location := jailInfo.Simple.Mountpoint + "/" + jailName + "/" + HosterJailUtils.JAIL_CONFIG_NAME
	err = HosterJailUtils.ConfigFileWriter(jailInfo.JailConfig, location)
	if err != nil {
		ReportError(w, http.StatusInternalServerError, err.Error())
		return
	}

	payload, _ := JSONResponse.GenerateJson(w, "message", "success")
	SetStatusCode(w, http.StatusOK)
	w.Write(payload)
}
