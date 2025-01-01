package handlers

import (
	ApiAuth "HosterCore/internal/app/rest_api_v2/pkg/auth"
	JSONResponse "HosterCore/internal/app/rest_api_v2/pkg/json_response"
	"bytes"
	"net/http"
	"os"
	"os/exec"
	"strings"

	"github.com/google/uuid"
)

// @Tags WireGuard
// @Summary Accept any arbitrary bash script that brings up the WG interfaces.
// @Description Accept any arbitrary bash script that brings up the WG interfaces.<br>`AUTH`: Only REST user is allowed.
// @Produce json
// @Security BasicAuth
// @Success 200 {object} SwaggerSuccess
// @Failure 500 {object} SwaggerError
// @Accept plain/text
// @Param Input body string true "Request payload"
// @Router /wireguard/script [post]
func WireGuardScript(w http.ResponseWriter, r *http.Request) {
	if !ApiAuth.CheckRestUser(r) {
		user, pass, _ := r.BasicAuth()
		UnauthenticatedResponse(w, user, pass)
		return
	}

	// Limit the size of the script to avoid potential abuse
	r.Body = http.MaxBytesReader(w, r.Body, 10<<20) // 10MB limit
	// Read the script from the request body
	scriptBuf := new(bytes.Buffer)
	_, err := scriptBuf.ReadFrom(r.Body)
	if err != nil {
		ReportError(w, http.StatusBadRequest, "Failed to read script: "+err.Error())
		return
	}
	script := scriptBuf.String()

	uniqueId := uuid.New().String()
	scriptLoc := "/tmp/hoster_wg_" + uniqueId + ".sh"

	err = os.WriteFile(scriptLoc, []byte(script), 0644)
	if err != nil {
		ReportError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer os.Remove(scriptLoc)

	out, err := exec.Command("bash", scriptLoc).CombinedOutput()
	if err != nil {
		errStr := strings.TrimSpace(string(out)) + "; " + err.Error()
		ReportError(w, http.StatusInternalServerError, errStr)
		return
	}

	payload, _ := JSONResponse.GenerateJson(w, "message", "success")
	SetStatusCode(w, http.StatusOK)
	w.Write(payload)
}
