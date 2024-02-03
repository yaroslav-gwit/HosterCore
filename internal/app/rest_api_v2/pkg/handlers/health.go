package handlers

import (
	ApiAuth "HosterCore/internal/app/rest_api_v2/pkg/auth"
	JSONResponse "HosterCore/internal/app/rest_api_v2/pkg/json_response"
	"net/http"

	"github.com/sirupsen/logrus"
)

// @Tags Health
// @Summary REST API server health status.
// @Description Simple function, that returns this REST API server health status.
// @Produce json
// @Success 200 {object} Models_SimpleSuccess
// @Router /health [get]
func HealthCheck(w http.ResponseWriter, r *http.Request) {
	payload, err := JSONResponse.GenerateJson(w, "message", "healthy")
	if err != nil {
		logrus.Error(err)
	}

	SetStatusCode(w, http.StatusOK)
	w.Write(payload)
}

// @Tags Health
// @Summary Check the regular user authentication.
// @Description Check the regular user authentication.
// @Produce json
// @Success 200 {object} Models_SimpleSuccess
// @Failure 500 {object} SwaggerError
// @Security BasicAuth
// @Router /health/ha-auth [get]
func HealthCheckHaAuth(w http.ResponseWriter, r *http.Request) {
	if ApiAuth.CheckHaUser(r) {
		payload, _ := JSONResponse.GenerateJson(w, "message", "success")
		SetStatusCode(w, http.StatusOK)
		w.Write(payload)
	} else {
		payload, _ := JSONResponse.GenerateJson(w, "message", "unauthorized")
		SetStatusCode(w, http.StatusUnauthorized)
		w.Write(payload)
	}
}

// @Tags Health
// @Summary Check the HA user authentication.
// @Description Check the HA user authentication.
// @Produce json
// @Success 200 {object} Models_SimpleSuccess
// @Failure 500 {object} SwaggerError
// @Security BasicAuth
// @Router /health/auth [get]
func HealthCheckAuth(w http.ResponseWriter, r *http.Request) {
	if ApiAuth.CheckRestUser(r) {
		payload, _ := JSONResponse.GenerateJson(w, "message", "success")
		SetStatusCode(w, http.StatusOK)
		w.Write(payload)
	} else {
		payload, _ := JSONResponse.GenerateJson(w, "message", "unauthorized")
		SetStatusCode(w, http.StatusUnauthorized)
		w.Write(payload)
	}
}
