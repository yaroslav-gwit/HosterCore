package main

import (
	JSONResponse "HosterCore/internal/app/rest_api_v2/json_response"
	"net/http"
)

func unauthenticatedResponse(w http.ResponseWriter) {
	w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
	payload, err := JSONResponse.GenerateJson(w, "error", "unauthorized")
	if err != nil {
		reportError(w, http.StatusInternalServerError, err.Error())
		return
	}

	setStatusCode(w, http.StatusUnauthorized)
	w.Write(payload)
}

func reportError(w http.ResponseWriter, httpStatusCode int, errorValue string) {
	setStatusCode(w, httpStatusCode)
	logApi.SetErrorMessage(errorValue)

	errPayload, _ := JSONResponse.GenerateJson(w, "error", errorValue)
	w.Write(errPayload)
}

func setStatusCode(w http.ResponseWriter, httpStatusCode int) {
	logApi.HttpStatusCode = httpStatusCode
	w.WriteHeader(httpStatusCode)
}
