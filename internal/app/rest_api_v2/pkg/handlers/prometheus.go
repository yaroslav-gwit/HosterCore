package handlers

import (
	ApiAuth "HosterCore/internal/app/rest_api_v2/pkg/auth"
	HosterPrometheus "HosterCore/internal/pkg/hoster/prometheus"
	"encoding/json"
	"net/http"
)

// @Tags Prometheus
// @Summary Generate a Prometheus autodiscovery JSON output for all Hoster VMs.
// @Description Generate a Prometheus autodiscovery JSON output for all Hoster VMs (this call will returns targets as DNS names).<br>`AUTH`: Only prometheus user is allowed.
// @Produce json
// @Security BasicAuth
// @Success 200 {object} []HosterPrometheus.PrometheusTarget{}
// @Failure 500 {object} SwaggerError
// @Router /prometheus/autodiscovery/vms [get]
func PrometheusAutoDiscoveryVms(w http.ResponseWriter, r *http.Request) {
	if !ApiAuth.CheckPrometheusUser(r) {
		user, pass, _ := r.BasicAuth()
		UnauthenticatedResponse(w, user, pass)
		return
	}

	prometheusTargets, err := HosterPrometheus.GenerateVmTargets(false)
	if err != nil {
		ReportError(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.Header().Add("content-type", "application/json")
	payload, err := json.Marshal(prometheusTargets)
	if err != nil {
		ReportError(w, http.StatusInternalServerError, err.Error())
		return
	}

	SetStatusCode(w, http.StatusOK)
	w.Write(payload)
}

// @Tags Prometheus
// @Summary Generate a Prometheus autodiscovery JSON output for all Hoster VMs.
// @Description Generate a Prometheus autodiscovery JSON output for all Hoster VMs (use IP addresses instead of DNS names).<br>`AUTH`: Only prometheus user is allowed.
// @Produce json
// @Security BasicAuth
// @Success 200 {object} []HosterPrometheus.PrometheusTarget{}
// @Failure 500 {object} SwaggerError
// @Router /prometheus/autodiscovery/vms/use-ips [get]
func PrometheusAutoDiscoveryVmsIps(w http.ResponseWriter, r *http.Request) {
	if !ApiAuth.CheckPrometheusUser(r) {
		user, pass, _ := r.BasicAuth()
		UnauthenticatedResponse(w, user, pass)
		return
	}

	prometheusTargets, err := HosterPrometheus.GenerateVmTargets(true)
	if err != nil {
		ReportError(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.Header().Add("content-type", "application/json")
	payload, err := json.Marshal(prometheusTargets)
	if err != nil {
		ReportError(w, http.StatusInternalServerError, err.Error())
		return
	}

	SetStatusCode(w, http.StatusOK)
	w.Write(payload)
}
