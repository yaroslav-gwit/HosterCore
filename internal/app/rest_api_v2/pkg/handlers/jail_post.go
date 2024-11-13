package handlers

import (
	ApiAuth "HosterCore/internal/app/rest_api_v2/pkg/auth"
	JSONResponse "HosterCore/internal/app/rest_api_v2/pkg/json_response"
	HosterJailUtils "HosterCore/internal/pkg/hoster/jail/utils"
	"encoding/json"
	"net/http"

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
