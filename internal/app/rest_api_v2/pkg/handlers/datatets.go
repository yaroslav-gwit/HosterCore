package handlers

import (
	ApiAuth "HosterCore/internal/app/rest_api_v2/pkg/auth"
	JSONResponse "HosterCore/internal/app/rest_api_v2/pkg/json_response"
	"HosterCore/internal/pkg/byteconversion"
	HosterHost "HosterCore/internal/pkg/hoster/host"
	HosterHostUtils "HosterCore/internal/pkg/hoster/host/utils"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

// @Tags Datasets
// @Summary Get active dataset list.
// @Description Get active dataset list.<br>`AUTH`: Only REST user is allowed.
// @Produce json
// @Security BasicAuth
// @Success 200 {object} []DatasetInfo
// @Failure 500 {object} SwaggerError
// @Router /dataset/all [get]
func DatasetList(w http.ResponseWriter, r *http.Request) {
	if !ApiAuth.CheckRestUser(r) {
		user, pass, _ := r.BasicAuth()
		UnauthenticatedResponse(w, user, pass)
		return
	}

	hostConf, err := HosterHost.GetHostConfig()
	if err != nil {
		ReportError(w, http.StatusInternalServerError, err.Error())
		return
	}

	info := []DatasetInfo{}
	for _, v := range hostConf.ActiveZfsDatasets {
		dsInfo, err := getDsInfo(v)
		if err != nil {
			continue
		}

		info = append(info, dsInfo)
	}

	payload, err := json.Marshal(info)
	if err != nil {
		ReportError(w, http.StatusInternalServerError, err.Error())
		return
	}

	SetStatusCode(w, http.StatusOK)
	w.Write(payload)
}

type DatasetInfo struct {
	Encrypted      bool   `json:"encrypted"`
	Mounted        bool   `json:"mounted"`
	Used           uint64 `json:"used"`
	Available      uint64 `json:"available"`
	Total          uint64 `json:"total"`
	UsedHuman      string `json:"used_human"`
	AvailableHuman string `json:"available_human"`
	TotalHuman     string `json:"total_human"`
	Name           string `json:"name"`
}

func getDsInfo(dsName string) (r DatasetInfo, e error) {
	reSpace := regexp.MustCompile(`\s+`)

	pool := strings.Split(dsName, "/")[0]
	out, err := exec.Command("zpool", "list", "-p", "-o", "name,size", pool).CombinedOutput()
	// Pool output example:
	//
	// zpool list -p -o name,size tank
	// NAME           SIZE
	//  [0]           [1]
	// tank          1992864825344
	if err != nil {
		e = fmt.Errorf("%s; %s", strings.TrimSpace(string(out)), err.Error())
		return
	}

	splitPool := strings.Split(string(out), "\n")
	poolValues := ""
	if len(splitPool) > 1 {
		poolValues = splitPool[1]
	} else {
		e = errors.New("could not find the pool")
		return
	}

	totalString := reSpace.Split(poolValues, -1)[1]
	total, err := strconv.ParseUint(totalString, 10, 64)
	if err != nil {
		e = err
		return
	}

	out, err = exec.Command("zfs", "list", "-p", "-o", "name,used,available,mounted,encryption", dsName).CombinedOutput()
	if err != nil {
		e = fmt.Errorf("%s; %s", strings.TrimSpace(string(out)), err.Error())
		return
	}

	// Encrypted example:
	// zfs list -p -o name,used,available,mounted,encryption tank/vm-encrypted
	//
	// NAME                USED          AVAIL        MOUNTED  ENCRYPTION
	// [0]                 [1]           [2]          [3]           [4]
	// tank/vm-encrypted  36383670272  1329563111424  yes      aes-256-gcm

	// Unencrypted example:
	// zfs list -p -o name,used,available,mounted,encryption tank/vm-unencrypted
	//
	// NAME                  USED       AVAIL       MOUNTED  ENCRYPTION
	// [0]                  [1]         [2]          [3]      [4]
	// tank/vm-unencrypted  98304   1329563111424    yes      off

	realValues := strings.Split(string(out), "\n")[1]
	split := reSpace.Split(realValues, -1)
	r.Name = split[0]

	// used, err := strconv.ParseUint(split[1], 10, 64)
	// if err != nil {
	// 	e = err
	// 	return
	// }
	// usedHuman := byteconversion.BytesToHuman(used)

	available, err := strconv.ParseUint(split[2], 10, 64)
	if err != nil {
		e = err
		return
	}
	availableHuman := byteconversion.BytesToHuman(available)

	r.Available = available
	r.AvailableHuman = availableHuman

	r.Used = total - available
	r.UsedHuman = byteconversion.BytesToHuman(r.Used)

	r.Total = total
	r.TotalHuman = byteconversion.BytesToHuman(r.Total)

	r.Mounted = split[3] == "yes"
	r.Encrypted = split[4] != "off"

	return
}

type DatasetEncryptionInput struct {
	Dataset  string `json:"dataset"`
	Password string `json:"password"`
}

// @Tags Datasets
// @Summary Unlock an encrypted dataset.
// @Description Unlock an encrypted dataset.<br>`AUTH`: Only REST user is allowed.
// @Produce json
// @Security BasicAuth
// @Success 200 {object} SwaggerSuccess
// @Failure 500 {object} SwaggerError
// @Param Input body DatasetEncryptionInput true "Request Payload"
// @Router /dataset/unlock [post]
func UnlockEncryptedDataset(w http.ResponseWriter, r *http.Request) {
	if !ApiAuth.CheckRestUser(r) {
		user, pass, _ := r.BasicAuth()
		UnauthenticatedResponse(w, user, pass)
		return
	}

	input := DatasetEncryptionInput{}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&input)
	if err != nil {
		ReportError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if len(input.Password) < 1 {
		ReportError(w, http.StatusBadRequest, "password is required")
		return
	}
	if len(input.Dataset) < 1 {
		ReportError(w, http.StatusBadRequest, "dataset is required")
		return
	}

	err = HosterHostUtils.UnlockEncryptedDataset(input.Dataset, input.Password)
	if err != nil {
		ReportError(w, http.StatusInternalServerError, err.Error())
		return
	}

	err = HosterHostUtils.ReloadDns()
	if err != nil {
		ReportError(w, http.StatusInternalServerError, err.Error())
		return
	}

	payload, err := JSONResponse.GenerateJson(w, "message", "success")
	if err != nil {
		ReportError(w, http.StatusInternalServerError, err.Error())
		return
	}

	SetStatusCode(w, http.StatusOK)
	w.Write(payload)
}
