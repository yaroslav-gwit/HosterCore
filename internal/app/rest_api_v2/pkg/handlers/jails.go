package handlers

import (
	JSONResponse "HosterCore/internal/app/rest_api_v2/pkg/json_response"
	HosterJail "HosterCore/internal/pkg/hoster/jail"
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
// @Summary List all Jail templates.
// @Description Get the list of all Jail templates.
// @Produce json
// @Success 200 {object} SwaggerStringList
// @Failure 500 {object} SwaggerError
// @Router /jail/templates [get]
func JailListTemplates(w http.ResponseWriter, r *http.Request) {
	templates, err := HosterJailUtils.ListTemplates()
	if err != nil {
		ReportError(w, http.StatusInternalServerError, err.Error())
		return
	}

	payload, _ := JSONResponse.GenerateJson(w, "message", templates)
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

// @Tags Jails
// @Summary Start a specific Jail.
// @Description Start a specific Jail using it's name as a parameter.
// @Produce json
// @Success 200 {object} SwaggerSuccess
// @Failure 500 {object} SwaggerError
// @Param jail_name path string true "Jail Name"
// @Router /jail/start/{jail_name} [post]
func JailStart(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	jailName := vars["jail_name"]

	err := HosterJail.Start(jailName)
	if err != nil {
		ReportError(w, http.StatusInternalServerError, err.Error())
		return
	}

	payload, _ := JSONResponse.GenerateJson(w, "message", "success")
	SetStatusCode(w, http.StatusOK)
	w.Write(payload)
}

// @Tags Jails
// @Summary Stop a specific Jail.
// @Description Stop a specific Jail using it's name as a parameter.
// @Produce json
// @Success 200 {object} SwaggerSuccess
// @Failure 500 {object} SwaggerError
// @Param jail_name path string true "Jail Name"
// @Router /jail/stop/{jail_name} [post]
func JailStop(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	jailName := vars["jail_name"]

	err := HosterJail.Stop(jailName)
	if err != nil {
		ReportError(w, http.StatusInternalServerError, err.Error())
		return
	}

	payload, _ := JSONResponse.GenerateJson(w, "message", "success")
	SetStatusCode(w, http.StatusOK)
	w.Write(payload)
}

// @Tags Jails
// @Summary Destroy a specific Jail.
// @Description Destroy a specific Jail using it's name as a parameter.<br>`DANGER` - destructive operation!
// @Produce json
// @Success 200 {object} SwaggerSuccess
// @Failure 500 {object} SwaggerError
// @Param jail_name path string true "Jail Name"
// @Router /jail/destroy/{jail_name} [delete]
func JailDestroy(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	jailName := vars["jail_name"]

	err := HosterJail.Destroy(jailName)
	if err != nil {
		ReportError(w, http.StatusInternalServerError, err.Error())
		return
	}

	payload, _ := JSONResponse.GenerateJson(w, "message", "success")
	SetStatusCode(w, http.StatusOK)
	w.Write(payload)
}
