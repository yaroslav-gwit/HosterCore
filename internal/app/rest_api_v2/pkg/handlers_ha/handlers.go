//go:build freebsd
// +build freebsd

package HandlersHA

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"time"

	"HosterCore/cmd"
	ApiAuth "HosterCore/internal/app/rest_api_v2/pkg/auth"
	RestApiConfig "HosterCore/internal/app/rest_api_v2/pkg/config"
	ErrorMappings "HosterCore/internal/app/rest_api_v2/pkg/error_mappings"
	"HosterCore/internal/app/rest_api_v2/pkg/handlers"
	JSONResponse "HosterCore/internal/app/rest_api_v2/pkg/json_response"
	HosterJailUtils "HosterCore/internal/pkg/hoster/jail/utils"
	HosterVmUtils "HosterCore/internal/pkg/hoster/vm/utils"
	zfsutils "HosterCore/internal/pkg/zfs_utils"
)

// @Tags HA
// @Summary Handle the HA node registration.
// @Description Handle the HA node registration.
// @Produce json
// @Success 200 {object} handlers.SwaggerSuccess
// @Failure 500 {object} handlers.SwaggerError
// @Security BasicAuth
// @Param Input body RestApiConfig.HaNode true "Request payload"
// @Router /ha/register [post]
func HandleRegistration(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if r := recover(); r != nil {
			errorValue := fmt.Sprintf("%s", r)
			// _ = exec.Command("logger", "-t", "HOSTER_HA_REST", "PANIC AVOIDED: handleHaRegistration(): "+errorValue).Run()
			internalLog.Warnf("PANIC AVOIDED: handleHaRegistration() %s", errorValue)
		}
	}()

	if !ApiAuth.CheckHaUser(r) {
		user, pass, _ := r.BasicAuth()
		handlers.UnauthenticatedResponse(w, user, pass)
		return
	}

	input := RestApiConfig.HaNode{}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&input)
	if err != nil {
		handlers.ReportError(w, http.StatusInternalServerError, ErrorMappings.CouldNotParseYourInput.String())
		return
	}

	hosterHaNode := HosterHaNode{}
	hosterHaNode.NodeInfo = input
	hosterHaNode.NodeInfo.Address = r.RemoteAddr
	hosterHaNode.LastPing = time.Now().Unix()

	modifyHostsDb(ModifyHostsDb{AddOrUpdate: true, Data: hosterHaNode}, &hostsDbLock)

	payload, _ := JSONResponse.GenerateJson(w, "message", "success")
	handlers.SetStatusCode(w, http.StatusOK)
	w.Write(payload)
}

// @Tags HA
// @Summary Handle the HA node ping.
// @Description Handle the HA node ping.
// @Produce json
// @Success 200 {object} handlers.SwaggerSuccess
// @Failure 500 {object} handlers.SwaggerError
// @Security BasicAuth
// @Param Input body RestApiConfig.HaNode true "Request payload"
// @Router /ha/ping [post]
func HandlePing(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if r := recover(); r != nil {
			errorValue := fmt.Sprintf("%s", r)
			// _ = exec.Command("logger", "-t", "HOSTER_HA_REST", "PANIC AVOIDED: handleHaPing(): "+errorValue).Run()
			internalLog.Warnf("PANIC AVOIDED: handleHaPing() %s", errorValue)
		}
	}()

	if !ApiAuth.CheckHaUser(r) {
		user, pass, _ := r.BasicAuth()
		handlers.UnauthenticatedResponse(w, user, pass)
		return
	}

	input := RestApiConfig.HaNode{}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&input)
	if err != nil {
		handlers.ReportError(w, http.StatusInternalServerError, ErrorMappings.CouldNotParseYourInput.String())
		return
	}

	hosterHaNode := HosterHaNode{}
	hosterHaNode.NodeInfo = input
	hosterHaNode.NodeInfo.Address = r.RemoteAddr
	hosterHaNode.LastPing = time.Now().Unix()

	modifyHostsDb(ModifyHostsDb{AddOrUpdate: true, Data: hosterHaNode}, &hostsDbLock)

	payload, _ := JSONResponse.GenerateJson(w, "message", "pong")
	handlers.SetStatusCode(w, http.StatusOK)
	w.Write(payload)
}

// @Tags HA
// @Summary Handle the HA graceful termination signal.
// @Description Handle the HA graceful termination signal.
// @Produce json
// @Success 200 {object} handlers.SwaggerSuccess
// @Failure 500 {object} handlers.SwaggerError
// @Security BasicAuth
// @Router /ha/terminate [post]
func HandleTerminate(w http.ResponseWriter, r *http.Request) {
	if !ApiAuth.CheckHaUser(r) {
		user, pass, _ := r.BasicAuth()
		handlers.UnauthenticatedResponse(w, user, pass)
		return
	}

	go func() {
		service := cmd.ApiProcessServiceInfo()
		// _ = exec.Command("logger", "-t", "HOSTER_HA_WATCHDOG", "INFO: received a remote terminating call").Run()
		internalLog.Warn("received a remote terminating call: TERMINATING WATCHDOG")
		_ = exec.Command("kill", "-SIGTERM", strconv.Itoa(service.HaWatchDogPid)).Run()
	}()

	go func() {
		time.Sleep(1500 * time.Millisecond)
		// _ = exec.Command("logger", "-t", "HOSTER_HA_REST", "INFO: received a remote terminating call").Run()
		internalLog.Warn("received a remote terminating call: TERMINATING REST API SERVICE")
		os.Exit(0)
	}()

	payload, _ := JSONResponse.GenerateJson(w, "message", "success")
	handlers.SetStatusCode(w, http.StatusOK)
	w.Write(payload)
}

type HaVm struct {
	VmName         string `json:"vm_name"`
	Live           bool   `json:"live"`
	LatestSnapshot string `json:"latest_snapshot"`
	ParentHost     string `json:"parent_host"`
	CurrentHost    string `json:"current_host"`
}

// @Tags HA
// @Summary Handle the HA enabled VM list.
// @Description Handle the HA enabled VM list.
// @Produce json
// @Success 200 {object} []HaVm
// @Failure 500 {object} handlers.SwaggerError
// @Security BasicAuth
// @Router /ha/vm-list [get]
func HandleVmList(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if r := recover(); r != nil {
			errorValue := fmt.Sprintf("%s", r)
			// _ = exec.Command("logger", "-t", "HOSTER_HA_REST", "PANIC AVOIDED: haVmsList(): "+errorValue).Run()
			internalLog.Warnf("PANIC AVOIDED: handleHaVmList(): %s", errorValue)
		}
	}()

	if !ApiAuth.CheckHaUser(r) {
		user, pass, _ := r.BasicAuth()
		handlers.UnauthenticatedResponse(w, user, pass)
		return
	}

	vms, err := HosterVmUtils.ListJsonApi()
	if err != nil {
		handlers.ReportError(w, http.StatusInternalServerError, err.Error())
		return
	}

	snaps, err := zfsutils.SnapshotListWithDescriptions()
	if err != nil {
		handlers.ReportError(w, http.StatusInternalServerError, err.Error())
		return
	}

	haVms := []HaVm{}
	for _, v := range vms {
		temp := HaVm{}
		if !v.Production {
			continue
		}

		temp.VmName = v.Name
		temp.Live = v.Running
		temp.ParentHost = v.ParentHost
		temp.CurrentHost = v.CurrentHost

		tempSnaps := []zfsutils.SnapshotInfo{}
		for _, vv := range snaps {
			if vv.Dataset == v.Simple.DsName {
				tempSnaps = append(tempSnaps, vv)
			}
		}
		temp.LatestSnapshot = tempSnaps[len(tempSnaps)-1].Name

		haVms = append(haVms, temp)
	}

	payload, err := json.Marshal(haVms)
	if err != nil {
		handlers.ReportError(w, http.StatusInternalServerError, err.Error())
		return
	}

	handlers.SetStatusCode(w, http.StatusOK)
	w.Write(payload)
}

type HaJail struct {
	VmName         string `json:"jail_name"`
	Live           bool   `json:"live"`
	LatestSnapshot string `json:"latest_snapshot"`
	ParentHost     string `json:"parent_host"`
	CurrentHost    string `json:"current_host"`
}

// @Tags HA
// @Summary Handle the HA enabled Jail list.
// @Description Handle the HA enabled Jail list.
// @Produce json
// @Success 200 {object} []HaJail
// @Failure 500 {object} handlers.SwaggerError
// @Security BasicAuth
// @Router /ha/jail-list [get]
func HandleJailList(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if r := recover(); r != nil {
			errorValue := fmt.Sprintf("%s", r)
			internalLog.Warnf("PANIC AVOIDED: handleHaJailList(): %s", errorValue)
		}
	}()

	if !ApiAuth.CheckHaUser(r) {
		user, pass, _ := r.BasicAuth()
		handlers.UnauthenticatedResponse(w, user, pass)
		return
	}

	jails, err := HosterJailUtils.ListJsonApi()
	if err != nil {
		handlers.ReportError(w, http.StatusInternalServerError, err.Error())
		return
	}

	snaps, err := zfsutils.SnapshotListWithDescriptions()
	if err != nil {
		handlers.ReportError(w, http.StatusInternalServerError, err.Error())
		return
	}

	haJails := []HaJail{}
	for _, v := range jails {
		temp := HaJail{}
		if !v.Production {
			continue
		}

		temp.VmName = v.Name
		temp.Live = v.Running
		temp.ParentHost = v.Parent
		temp.CurrentHost = v.CurrentHost

		tempSnaps := []zfsutils.SnapshotInfo{}
		for _, vv := range snaps {
			if vv.Dataset == v.Simple.DsName {
				tempSnaps = append(tempSnaps, vv)
			}
		}
		temp.LatestSnapshot = tempSnaps[len(tempSnaps)-1].Name

		haJails = append(haJails, temp)
	}

	payload, err := json.Marshal(haJails)
	if err != nil {
		handlers.ReportError(w, http.StatusInternalServerError, err.Error())
		return
	}

	handlers.SetStatusCode(w, http.StatusOK)
	w.Write(payload)
}
