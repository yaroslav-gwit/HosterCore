package handlers

import (
	ApiAuth "HosterCore/internal/app/rest_api_v2/pkg/auth"
	JSONResponse "HosterCore/internal/app/rest_api_v2/pkg/json_response"
	"net/http"
)

// @Tags Health
// @Summary REST API server health status.
// @Description Simple function, that returns this REST API server health status.
// @Produce json
// @Success 200 {object} Models_SimpleSuccess
// @Router /health [get]
func HealthCheck(w http.ResponseWriter, r *http.Request) {
	payload, _ := JSONResponse.GenerateJson(w, "message", "healthy")
	SetStatusCode(w, http.StatusOK)
	w.Write(payload)
}

// @Tags Health
// @Summary Check the `regular` user authentication.
// @Description Check the `regular` user authentication.
// @Produce json
// @Success 200 {object} Models_SimpleSuccess
// @Failure 500 {object} SwaggerError
// @Security BasicAuth
// @Router /health/auth/regular [get]
func HealthCheckRegularAuth(w http.ResponseWriter, r *http.Request) {
	auth := ApiAuth.CheckRestUser(r)
	if auth {
		payload, _ := JSONResponse.GenerateJson(w, "message", "success")
		SetStatusCode(w, http.StatusOK)
		w.Write(payload)
	} else {
		user, pass, _ := r.BasicAuth()
		UnauthenticatedResponse(w, user, pass)
	}
}

// @Tags Health
// @Summary Check the `HA` user authentication.
// @Description Check the `HA` user authentication.
// @Produce json
// @Success 200 {object} Models_SimpleSuccess
// @Failure 500 {object} SwaggerError
// @Security BasicAuth
// @Router /health/auth/ha [get]
func HealthCheckHaAuth(w http.ResponseWriter, r *http.Request) {
	auth := ApiAuth.CheckHaUser(r)
	if auth {
		payload, _ := JSONResponse.GenerateJson(w, "message", "success")
		SetStatusCode(w, http.StatusOK)
		w.Write(payload)
	} else {
		user, pass, _ := r.BasicAuth()
		UnauthenticatedResponse(w, user, pass)
	}
}

// @Tags Health
// @Summary Check `any` user authentication.
// @Description Check if `any` of the two users can log in. Useful for the routes which are required by both users: regular and HA.
// @Produce json
// @Success 200 {object} Models_SimpleSuccess
// @Failure 500 {object} SwaggerError
// @Security BasicAuth
// @Router /health/auth/any [get]
func HealthCheckAnyAuth(w http.ResponseWriter, r *http.Request) {
	auth := ApiAuth.CheckAnyUser(r)
	if auth {
		payload, _ := JSONResponse.GenerateJson(w, "message", "success")
		SetStatusCode(w, http.StatusOK)
		w.Write(payload)
	} else {
		user, pass, _ := r.BasicAuth()
		UnauthenticatedResponse(w, user, pass)
	}
}
