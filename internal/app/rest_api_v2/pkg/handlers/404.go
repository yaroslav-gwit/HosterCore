package handlers

import (
	JSONResponse "HosterCore/internal/app/rest_api_v2/pkg/json_response"
	"net/http"
)

func NotFoundHandler(w http.ResponseWriter, r *http.Request) {
	payload, _ := JSONResponse.GenerateJson(w, "message", "404")

	log.SetErrorMessage("404")
	SetStatusCode(w, http.StatusNotFound)
	w.Write(payload)
}
