package handlers

import (
	ErrorMappings "HosterCore/internal/app/rest_api_v2/pkg/error_mappings"
	JSONResponse "HosterCore/internal/app/rest_api_v2/pkg/json_response"
	MiddlewareLogging "HosterCore/internal/app/rest_api_v2/pkg/middleware/logging"
	"encoding/json"
	"net/http"

	"github.com/sirupsen/logrus"
)

var log *MiddlewareLogging.Log

func init() {
	log = MiddlewareLogging.Configure(logrus.DebugLevel)
}

func SetStatusCode(w http.ResponseWriter, httpStatusCode int) {
	w.Header().Add("Access-Control-Allow-Methods", "*")
	w.Header().Add("Access-Control-Allow-Origin", "*")
	log.HttpStatusCode = httpStatusCode
	w.WriteHeader(httpStatusCode)
}

func ReportError(w http.ResponseWriter, httpStatusCode int, errorValue string) {
	log.SetErrorMessage(errorValue)

	swaggerErr := SwaggerError{}
	errID := ErrorMappings.ValueLookup(errorValue)
	swaggerErr.ErrorID = int(errID)
	swaggerErr.ErrorValue = errorValue

	payload, _ := json.Marshal(swaggerErr)
	w.Header().Add("Content-Type", "application/json")
	SetStatusCode(w, httpStatusCode)
	w.Write(payload)
}

func UnauthenticatedResponse(w http.ResponseWriter) {
	w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
	payload, err := JSONResponse.GenerateJson(w, "error", "unauthorized")
	if err != nil {
		ReportError(w, http.StatusUnauthorized, err.Error())
		return
	}

	SetStatusCode(w, http.StatusUnauthorized)
	w.Write(payload)
}

func SetLogConfig(l *MiddlewareLogging.Log) {
	log = l
}
