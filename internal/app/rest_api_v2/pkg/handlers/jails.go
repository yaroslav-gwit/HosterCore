package handlers

import (
	ApiAuth "HosterCore/internal/app/rest_api_v2/pkg/auth"
	JSONResponse "HosterCore/internal/app/rest_api_v2/pkg/json_response"
	HosterJail "HosterCore/internal/pkg/hoster/jail"
	HosterJailUtils "HosterCore/internal/pkg/hoster/jail/utils"
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
)

// @Tags Jails
// @Summary List all Jails.
// @Description Get the list of all Jails, including the information about them.<br>`AUTH`: Both users are allowed.
// @Produce json
// @Security BasicAuth
// @Success 200 {object} []HosterJailUtils.JailApi
// @Failure 500 {object} SwaggerError
// @Router /jail/all [get]
func JailList(w http.ResponseWriter, r *http.Request) {
	if !ApiAuth.CheckAnyUser(r) {
		user, pass, _ := r.BasicAuth()
		UnauthenticatedResponse(w, user, pass)
		return
	}

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
// @Summary List all Jails (cached version).
// @Description Get the list of all Jails, including the information about them (cached version).<br>`AUTH`: Both users are allowed.
// @Produce json
// @Security BasicAuth
// @Success 200 {object} []HosterJailUtils.JailApi
// @Failure 500 {object} SwaggerError
// @Router /jail/all/cache [get]
func JailListCache(w http.ResponseWriter, r *http.Request) {
	if !ApiAuth.CheckAnyUser(r) {
		user, pass, _ := r.BasicAuth()
		UnauthenticatedResponse(w, user, pass)
		return
	}

	jails, err := HosterJailUtils.ReadCache()
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
// @Description Get the list of all Jail templates.<br>`AUTH`: Only `rest` user is allowed.
// @Produce json
// @Security BasicAuth
// @Success 200 {array} string
// @Failure 500 {object} SwaggerError
// @Router /jail/template/list [get]
func JailListTemplates(w http.ResponseWriter, r *http.Request) {
	if !ApiAuth.CheckRestUser(r) {
		user, pass, _ := r.BasicAuth()
		UnauthenticatedResponse(w, user, pass)
		return
	}

	templates, err := HosterJailUtils.ListTemplates()
	if err != nil {
		ReportError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// payload, _ := JSONResponse.GenerateJson(w, "message", templates)
	payload, err := json.Marshal(templates)
	if err != nil {
		ReportError(w, http.StatusInternalServerError, err.Error())
		return
	}

	SetStatusCode(w, http.StatusOK)
	w.Write(payload)
}

// @Tags Jails
// @Summary Get Jail info.
// @Description Get Jail info.<br>`AUTH`: Both users are allowed.
// @Produce json
// @Security BasicAuth
// @Success 200 {object} HosterJailUtils.JailApi
// @Failure 500 {object} SwaggerError
// @Param jail_name path string true "Jail Name"
// @Router /jail/info/{jail_name} [get]
func JailInfo(w http.ResponseWriter, r *http.Request) {
	if !ApiAuth.CheckAnyUser(r) {
		user, pass, _ := r.BasicAuth()
		UnauthenticatedResponse(w, user, pass)
		return
	}

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

	w.Header().Set("Content-Type", "application/json")
	SetStatusCode(w, http.StatusOK)
	w.Write(payload)
}

// @Tags Jails
// @Summary Start a specific Jail.
// @Description Start a specific Jail using it's name as a parameter.<br>`AUTH`: Both users are allowed.
// @Produce json
// @Security BasicAuth
// @Success 200 {object} SwaggerSuccess
// @Failure 500 {object} SwaggerError
// @Param jail_name path string true "Jail Name"
// @Router /jail/start/{jail_name} [post]
func JailStart(w http.ResponseWriter, r *http.Request) {
	if !ApiAuth.CheckAnyUser(r) {
		user, pass, _ := r.BasicAuth()
		UnauthenticatedResponse(w, user, pass)
		return
	}

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
// @Description Stop a specific Jail using it's name as a parameter.<br>`AUTH`: Both users are allowed.
// @Produce json
// @Security BasicAuth
// @Success 200 {object} SwaggerSuccess
// @Failure 500 {object} SwaggerError
// @Param jail_name path string true "Jail Name"
// @Router /jail/stop/{jail_name} [post]
func JailStop(w http.ResponseWriter, r *http.Request) {
	if !ApiAuth.CheckAnyUser(r) {
		user, pass, _ := r.BasicAuth()
		UnauthenticatedResponse(w, user, pass)
		return
	}

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
// @Description `DANGER` - destructive operation!<br><br>Destroy a specific Jail using it's name as a parameter.<br>`AUTH`: Only `rest` user is allowed.
// @Produce json
// @Security BasicAuth
// @Success 200 {object} SwaggerSuccess
// @Failure 500 {object} SwaggerError
// @Param jail_name path string true "Jail Name"
// @Router /jail/destroy/{jail_name} [delete]
func JailDestroy(w http.ResponseWriter, r *http.Request) {
	if !ApiAuth.CheckRestUser(r) {
		user, pass, _ := r.BasicAuth()
		UnauthenticatedResponse(w, user, pass)
		return
	}

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

// @Tags Jails
// @Summary Deploy a new Jail.
// @Description Deploy a new Jail using a set of defined parameters.<br>`AUTH`: Only `rest` user is allowed.
// @Produce json
// @Security BasicAuth
// @Success 200 {object} SwaggerSuccess
// @Failure 500 {object} SwaggerError
// @Param Input body HosterJail.DeployInput true "Request payload"
// @Router /jail/deploy [post]
func JailDeploy(w http.ResponseWriter, r *http.Request) {
	if !ApiAuth.CheckRestUser(r) {
		user, pass, _ := r.BasicAuth()
		UnauthenticatedResponse(w, user, pass)
		return
	}

	input := HosterJail.DeployInput{}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&input)
	if err != nil {
		ReportError(w, http.StatusInternalServerError, err.Error())
		return
	}

	err = HosterJail.Deploy(input)
	if err != nil {
		ReportError(w, http.StatusInternalServerError, err.Error())
		return
	}

	payload, _ := JSONResponse.GenerateJson(w, "message", "success")
	SetStatusCode(w, http.StatusOK)
	w.Write(payload)
}

// @Tags Jails
// @Summary Clone the Jail.
// @Description Clone the Jail using it's name, and optionally specify the snapshot name to be used for cloning.<br>`AUTH`: Only `rest` user is allowed.
// @Produce json
// @Security BasicAuth
// @Success 200 {object} SwaggerSuccess
// @Failure 500 {object} SwaggerError
// @Param Input body JailCloneInput true "Request payload"
// @Router /jail/clone [post]
func JailClone(w http.ResponseWriter, r *http.Request) {
	if !ApiAuth.CheckRestUser(r) {
		user, pass, _ := r.BasicAuth()
		UnauthenticatedResponse(w, user, pass)
		return
	}

	input := JailCloneInput{}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&input)
	if err != nil {
		ReportError(w, http.StatusInternalServerError, err.Error())
		return
	}

	err = HosterJail.Clone(input.JailName, input.NewJailName, input.SnapshotName)
	if err != nil {
		ReportError(w, http.StatusInternalServerError, err.Error())
		return
	}

	payload, _ := JSONResponse.GenerateJson(w, "message", "success")
	SetStatusCode(w, http.StatusOK)
	w.Write(payload)
}

// @Tags Jails
// @Summary Get README.MD for a particular Jail.
// @Description Get README.MD for a particular Jail.<br>`AUTH`: Only `rest` user is allowed.
// @Produce json
// @Security BasicAuth
// @Success 200 {object} SwaggerSuccess
// @Failure 500 {object} SwaggerError
// @Param jail_name path string true "Jail Name"
// @Router /jail/readme/{jail_name} [get]
func JailGetReadme(w http.ResponseWriter, r *http.Request) {
	if !ApiAuth.CheckRestUser(r) {
		user, pass, _ := r.BasicAuth()
		UnauthenticatedResponse(w, user, pass)
		return
	}

	vars := mux.Vars(r)
	jailName := vars["jail_name"]

	readme, err := HosterJail.GetReadme(jailName)
	if err != nil {
		ReportError(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.Header().Add("Content-Type", "text/plain")
	SetStatusCode(w, http.StatusOK)
	w.Write([]byte(readme))
}

// @Tags Jails
// @Summary Get a list of active shells for a specific Jail.
// @Description Get a list of active shells for a specific Jail.<br>`AUTH`: Only `rest` user is allowed.
// @Produce json
// @Security BasicAuth
// @Success 200 {object} JailShells{}
// @Failure 500 {object} SwaggerError
// @Param jail_name path string true "Jail Name"
// @Router /jail/get/shells/{jail_name} [get]
func JailGetShells(w http.ResponseWriter, r *http.Request) {
	if !ApiAuth.CheckRestUser(r) {
		user, pass, _ := r.BasicAuth()
		UnauthenticatedResponse(w, user, pass)
		return
	}

	vars := mux.Vars(r)
	jailName := vars["jail_name"]

	var err error
	output := JailShells{}
	output.AvailableShells, err = HosterJailUtils.GetJailShells(jailName)
	if err != nil {
		ReportError(w, http.StatusInternalServerError, err.Error())
		return
	}

	payload, err := json.Marshal(output)
	if err != nil {
		ReportError(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	SetStatusCode(w, http.StatusOK)
	w.Write(payload)
}
