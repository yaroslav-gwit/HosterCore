package main

import (
	"HosterCore/cmd"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

type NodeInfoStruct struct {
	Hostname         string `json:"hostname"`
	Protocol         string `json:"protocol"`
	Address          string `json:"address"`
	Port             string `json:"port"`
	User             string `json:"user"`
	Password         string `json:"password"`
	FailOverStrategy string `json:"failover_strategy"`
	FailOverTime     int64  `json:"failover_time"`
	BackupNode       bool   `json:"backup_node"`
	StartupTime      int64  `json:"startup_time"`
	Registered       bool   `json:"registered"`
	TimesFailed      int    `json:"times_failed"`
}

type HosterHaNodeStruct struct {
	LastPing int64          `json:"last_ping"`
	NodeInfo NodeInfoStruct `json:"node_info"`
}

type HaConfigJsonStruct struct {
	NodeType         string           `json:"node_type"`
	FailOverStrategy string           `json:"failover_strategy"`
	FailOverTime     int64            `json:"failover_time"`
	BackupNode       bool             `json:"backup_node"`
	Candidates       []NodeInfoStruct `json:"candidates"`
	StartupTime      int64            `json:"startup_time"`
}

var haHostsDb []HosterHaNodeStruct
var haConfig HaConfigJsonStruct

// var hostsDbLock sync.RWMutex

type ModifyHostsDbStruct struct {
	addOrUpdate bool
	remove      bool
	data        HosterHaNodeStruct
}

var haMode bool
var debugMode bool

var user = "admin"
var password = "123456"
var port = 3000
var protocol = "http"

func init() {
	portEnv := os.Getenv("REST_API_PORT")
	userEnv := os.Getenv("REST_API_USER")
	passwordEnv := os.Getenv("REST_API_PASSWORD")

	var err error
	if len(portEnv) > 0 {
		port, err = strconv.Atoi(portEnv)
		if err != nil {
			log.Fatal("please make sure port is an integer!")
		}
	}
	if len(userEnv) > 0 {
		user = userEnv
	}
	if len(passwordEnv) > 0 {
		password = passwordEnv
	}

	haModeEnv := os.Getenv("REST_API_HA_MODE")
	if len(haModeEnv) > 0 {
		haMode = true
	} else {
		_ = exec.Command("logger", "-t", "HOSTER_REST", "INFO: STARING REST API SERVER IN REGULAR (NON-HA) MODE").Run()
		return
	}

	debugModeEnv := os.Getenv("REST_API_HA_DEBUG")
	if len(debugModeEnv) > 0 {
		debugMode = true
	}

	if debugMode {
		haModeString := []byte("DEBUG")
		_ = os.Remove("/var/run/hoster_rest_api.mode")
		err := os.WriteFile("/var/run/hoster_rest_api.mode", haModeString, 0644)
		if err != nil {
			_ = exec.Command("logger", "-t", "HOSTER_REST", "ERROR: could not write the /var/run/hoster_rest_api.mode file").Run()
		}
	} else {
		haModeString := []byte("PRODUCTION")
		_ = os.Remove("/var/run/hoster_rest_api.mode")
		err := os.WriteFile("/var/run/hoster_rest_api.mode", haModeString, 0644)
		if err != nil {
			_ = exec.Command("logger", "-t", "HOSTER_REST", "ERROR: could not write the /var/run/hoster_rest_api.mode file").Run()
		}
	}

	file, _ := os.ReadFile("/opt/hoster-core/config_files/ha_config.json")
	err = json.Unmarshal(file, &haConfig)
	if err != nil {
		_ = exec.Command("logger", "-t", "HOSTER_HA", "PANIC: could not parse the HA configuration file: "+err.Error()).Run()
		panic("Cannot parse the HA configuration file: " + err.Error())
	}

	if haConfig.FailOverTime < 1 {
		haConfig.FailOverTime = 60
	}
	_ = exec.Command("logger", "-t", "HOSTER_HA_REST", "INFO: node failover time is: "+strconv.Itoa(int(haConfig.FailOverTime))+" seconds").Run()

	haConfig.StartupTime = time.Now().UnixMilli()

	// Candidates and workers
	go registerNode()
	go trackCandidatesOnline()
	go sendPing()

	// Candidates only
	for _, v := range haConfig.Candidates {
		if v.Hostname == cmd.GetHostName() {
			go trackManager()
			go removeOfflineNodes()
		}
	}
}

var candidatesRegistered = 0
var clusterInitialized = false
var iAmManager = false
var myHostname = cmd.GetHostName()

func trackCandidatesOnline() {
	defer func() {
		if r := recover(); r != nil {
			errorValue := fmt.Sprintf("%s", r)
			_ = exec.Command("logger", "-t", "HOSTER_HA_REST", "PANIC AVOIDED: trackCandidatesOnline(): "+errorValue).Run()
		}
	}()

	for {
		candidatesRegistered = 0

		for _, v := range haConfig.Candidates {
			if v.Registered {
				candidatesRegistered += 1
			}
		}

		if !clusterInitialized && candidatesRegistered >= 3 {
			clusterInitialized = true
		}

		if clusterInitialized && candidatesRegistered < 2 {
			_ = exec.Command("logger", "-t", "HOSTER_HA_REST", "EMERG: candidatesRegistered has gone below 2, initiating self fencing").Run()
			os.Exit(0)
		}

		time.Sleep(time.Second * 10)
	}
}

func trackManager() {
	defer func() {
		if r := recover(); r != nil {
			errorValue := fmt.Sprintf("%s", r)
			_ = exec.Command("logger", "-t", "HOSTER_HA_REST", "PANIC AVOIDED: trackManager(): "+errorValue).Run()
		}
	}()

	for {
		if clusterInitialized {
			var copyHostsDb = readHostsDb()
			var filteredCandidates []HosterHaNodeStruct

			sort.Slice(copyHostsDb, func(i int, j int) bool {
				return copyHostsDb[i].NodeInfo.StartupTime < copyHostsDb[j].NodeInfo.StartupTime
			})

			for _, host := range copyHostsDb {
				for _, candidate := range haConfig.Candidates {
					if host.NodeInfo.Hostname == candidate.Hostname {
						if candidate.Registered {
							filteredCandidates = append(filteredCandidates, host)
						}
						break
					}
				}
			}

			if filteredCandidates[0].NodeInfo.Hostname == myHostname {
				if !iAmManager {
					_ = exec.Command("logger", "-t", "HOSTER_HA_REST", "INFO: becoming a new cluster manager").Run()
					iAmManager = true
				}
			} else {
				if iAmManager {
					_ = exec.Command("logger", "-t", "HOSTER_HA_REST", "WARN: stepping down as a cluster manager").Run()
					iAmManager = false
				}
			}

			time.Sleep(time.Second * 4)
		}
	}
}

func registerNode() {
	defer func() {
		if r := recover(); r != nil {
			errorValue := fmt.Sprintf("%s", r)
			_ = exec.Command("logger", "-t", "HOSTER_HA_REST", "PANIC AVOIDED: registerNode(): "+errorValue).Run()
		}
	}()

	for {
		if candidatesRegistered >= 3 {
			time.Sleep(time.Second * 10)
			continue
		}

		time.Sleep(time.Second * 10)
		for i, v := range haConfig.Candidates {
			if v.Registered {
				continue
			}
			host := NodeInfoStruct{}
			host.Hostname = cmd.GetHostName()
			host.FailOverStrategy = haConfig.FailOverStrategy
			host.User = user
			host.Password = password
			host.Port = strconv.Itoa(port)
			host.Protocol = protocol
			host.FailOverStrategy = haConfig.FailOverStrategy
			host.FailOverTime = haConfig.FailOverTime
			host.StartupTime = haConfig.StartupTime
			host.BackupNode = haConfig.BackupNode

			jsonPayload, _ := json.Marshal(host)
			payload := strings.NewReader(string(jsonPayload))
			url := v.Protocol + "://" + v.Address + ":" + v.Port + "/api/v1/ha/register"

			req, err := http.NewRequest("POST", url, payload)
			if err != nil {
				_ = exec.Command("logger", "-t", "HOSTER_HA_REST", "ERROR: could not form the /register request: "+err.Error()).Run()
				time.Sleep(time.Second * 10)
				continue
			}

			auth := v.User + ":" + v.Password
			authEncoded := base64.StdEncoding.EncodeToString([]byte(auth))
			req.Header.Add("Content-Type", "application/json")
			req.Header.Add("Authorization", "Basic "+authEncoded)

			res, err := http.DefaultClient.Do(req)
			if err != nil {
				_ = exec.Command("logger", "-t", "HOSTER_HA_REST", "WARN: could not join the candidate: "+err.Error()).Run()
				time.Sleep(time.Second * 30)
				continue
			} else {
				_ = exec.Command("logger", "-t", "HOSTER_HA_REST", "SUCCESS: joined the candidate: "+v.Hostname).Run()
				haConfig.Candidates[i].Registered = true
				haConfig.Candidates[i].StartupTime = haConfig.StartupTime
				req.Body.Close()
				res.Body.Close()
			}
		}
	}
}

func sendPing() {
	defer func() {
		if r := recover(); r != nil {
			errorValue := fmt.Sprintf("%s", r)
			_ = exec.Command("logger", "-t", "HOSTER_HA_REST", "PANIC AVOIDED: sendPing(): "+errorValue).Run()
		}
	}()

	for {
		if !clusterInitialized {
			time.Sleep(time.Second * 10)
			continue
		}

		var wg sync.WaitGroup
		wg.Add(len(haConfig.Candidates))

		for i, v := range haConfig.Candidates {
			if !v.Registered {
				continue
			}

			go func(i int, v NodeInfoStruct) {
				defer func() {
					r := recover()
					if r != nil {
						errorValue := fmt.Sprintf("%s", r)
						_ = exec.Command("logger", "-t", "HOSTER_HA_REST", "PANIC AVOIDED: sendPingMainGoRoutine(): "+errorValue).Run()
					}
					wg.Done()
				}()

				host := NodeInfoStruct{}
				host.Hostname = cmd.GetHostName()
				host.StartupTime = haConfig.StartupTime

				jsonPayload, err := json.Marshal(host)
				if err != nil {
					_ = exec.Command("logger", "-t", "HOSTER_HA_REST", "ERR: pingGoRoutineJSONPAYLOAD(): "+err.Error()).Run()
					return
				}

				payload := strings.NewReader(string(jsonPayload))
				url := v.Protocol + "://" + v.Address + ":" + v.Port + "/api/v1/ha/ping"
				req, err := http.NewRequest("POST", url, payload)
				if err != nil {
					_ = exec.Command("logger", "-t", "HOSTER_HA_REST", "ERR: pingGoRoutineREQ(): "+err.Error()).Run()
					return
				}

				auth := v.User + ":" + v.Password
				authEncoded := base64.StdEncoding.EncodeToString([]byte(auth))
				req.Header.Add("Content-Type", "application/json")
				req.Header.Add("Authorization", "Basic "+authEncoded)
				resp, err := http.DefaultClient.Do(req)
				if err != nil {
					_ = exec.Command("logger", "-t", "HOSTER_HA_REST", "WARN: failed to ping the candidate node: "+err.Error()).Run()
					haConfig.Candidates[i].TimesFailed += 1
					if haConfig.Candidates[i].TimesFailed >= 3 {
						haConfig.Candidates[i].Registered = false
					}
				} else {
					if haConfig.Candidates[i].TimesFailed > 0 {
						haConfig.Candidates[i].TimesFailed -= 1
					}
					haConfig.Candidates[i].Registered = true
					// Close the request and response body to release resources
					req.Body.Close()
					resp.Body.Close()
				}
			}(i, v)
		}

		time.Sleep(time.Second * 2)
	}
}

func removeOfflineNodes() {
	defer func() {
		if r := recover(); r != nil {
			errorValue := fmt.Sprintf("%s", r)
			_ = exec.Command("logger", "-t", "HOSTER_HA_REST", "PANIC AVOIDED: removeOfflineNodes(): "+errorValue).Run()
		}
	}()

	for {
		hostsDbCopy := readHostsDb()
		for _, v := range hostsDbCopy {
			if time.Now().Unix() > v.LastPing+v.NodeInfo.FailOverTime {
				if len(v.NodeInfo.Hostname) > 0 {
					failoverHostVms(v)
					modifyHostsDb(ModifyHostsDbStruct{data: v, remove: true})
					_ = exec.Command("logger", "-t", "HOSTER_HA_REST", "WARN: host has gone offline: "+v.NodeInfo.Hostname).Run()
				}
			}
		}
		time.Sleep(time.Second * 4)
	}
}

func failoverHostVms(haNode HosterHaNodeStruct) {
	defer func() {
		if r := recover(); r != nil {
			errorValue := fmt.Sprintf("%s", r)
			_ = exec.Command("logger", "-t", "HOSTER_HA_REST", "PANIC AVOIDED: failoverHostVms(): "+errorValue).Run()
		}
	}()

	if !iAmManager {
		time.Sleep(time.Second * 3)
		return
	}

	haVms := []HaVm{}
	hostsDbCopy := readHostsDb()
	for _, v := range hostsDbCopy {
		haVmsTemp := []HaVm{}

		// Skip the failed node (passed via function parameters, and not offline-d yet)
		if v.NodeInfo.Hostname == haNode.NodeInfo.Hostname {
			continue
		}
		// Skip if the node in question is a backup host, participating purely for quorum purposes
		if v.NodeInfo.BackupNode {
			continue
		}

		url := v.NodeInfo.Protocol + "://" + v.NodeInfo.Address + ":" + v.NodeInfo.Port + "/api/v1/ha/vms"
		req, _ := http.NewRequest("GET", url, nil)
		auth := v.NodeInfo.User + ":" + v.NodeInfo.Password
		authEncoded := base64.StdEncoding.EncodeToString([]byte(auth))
		req.Header.Add("Authorization", "Basic "+authEncoded)
		res, err := http.DefaultClient.Do(req)
		if err != nil {
			_ = exec.Command("logger", "-t", "HOSTER_HA_REST", "ERROR: line 345: "+err.Error()).Run()
			continue
		}

		defer res.Body.Close()
		body, _ := io.ReadAll(res.Body)

		err = json.Unmarshal(body, &haVmsTemp)
		if err != nil {
			_ = exec.Command("logger", "-t", "HOSTER_HA_REST", "ERROR: line 354: "+err.Error()).Run()
			continue
		}

		for _, vv := range haVmsTemp {
			if vv.ParentHost == haNode.NodeInfo.Hostname {
				haVms = append(haVms, vv)
			}
		}
	}

	sort.Slice(haVms, func(i int, j int) bool {
		return haVms[i].LatestSnapshot < haVms[j].LatestSnapshot
	})

	uniqueHaVms := []HaVm{}
	for _, v := range haVms {
		vmExists := false
		for _, vv := range uniqueHaVms {
			if v.VmName == vv.VmName {
				vmExists = true
			}
		}
		if !vmExists {
			uniqueHaVms = append(uniqueHaVms, v)
		}
	}

	for _, v := range uniqueHaVms {
		if debugMode {
			_ = exec.Command("logger", "-t", "HOSTER_HA_REST", "WARN: MOVING VM: "+v.VmName+" FROM offline parent: "+v.ParentHost+" TO: "+v.CurrentHost).Run()
			continue
		}

		for _, vv := range hostsDbCopy {
			if vv.NodeInfo.Hostname == v.CurrentHost {
				time.Sleep(1500 * time.Millisecond)
				_ = exec.Command("logger", "-t", "HOSTER_HA_REST", "WARN: MOVING VM: "+v.VmName+" FROM offline parent: "+v.ParentHost+" TO: "+v.CurrentHost).Run()

				auth := vv.NodeInfo.User + ":" + vv.NodeInfo.Password
				authEncoded := base64.StdEncoding.EncodeToString([]byte(auth))

				// Use failover strategy to failover the VM
				if vv.NodeInfo.FailOverStrategy == "cireset" || vv.NodeInfo.FailOverStrategy == "ci-reset" {
					url := vv.NodeInfo.Protocol + "://" + vv.NodeInfo.Address + ":" + vv.NodeInfo.Port + "/api/v1/vm/cireset"
					payload := strings.NewReader(`{ "name": "` + v.VmName + `" }`)
					req, _ := http.NewRequest("POST", url, payload)

					req.Header.Add("Content-Type", "application/json")
					req.Header.Add("Authorization", "Basic "+authEncoded)
					res, err := http.DefaultClient.Do(req)
					if res.StatusCode != 200 {
						_ = err
						_ = exec.Command("logger", "-t", "HOSTER_HA_REST", "ERROR: CIRESET FAILED FOR THE VM: "+v.VmName+" ON: "+v.CurrentHost).Run()
						continue
					}
					req.Body.Close()
					res.Body.Close()
				} else if vv.NodeInfo.FailOverStrategy == "changeparent" || vv.NodeInfo.FailOverStrategy == "change-parent" {
					url := vv.NodeInfo.Protocol + "://" + vv.NodeInfo.Address + ":" + vv.NodeInfo.Port + "/api/v1/vm/change-parent"
					payload := strings.NewReader(`{ "name": "` + v.VmName + `" }`)
					req, _ := http.NewRequest("POST", url, payload)

					req.Header.Add("Content-Type", "application/json")
					req.Header.Add("Authorization", "Basic "+authEncoded)
					res, err := http.DefaultClient.Do(req)
					if res.StatusCode != 200 {
						_ = err
						_ = exec.Command("logger", "-t", "HOSTER_HA_REST", "ERROR: CHANGE PARENT FAILED FOR THE VM: "+v.VmName+" ON: "+v.CurrentHost).Run()
						continue
					}
					req.Body.Close()
					res.Body.Close()
				}

				// Start VM on a new host
				url := vv.NodeInfo.Protocol + "://" + vv.NodeInfo.Address + ":" + vv.NodeInfo.Port + "/api/v1/vm/start"
				payload := strings.NewReader(`{ "name": "` + v.VmName + `" }`)
				req, _ := http.NewRequest("POST", url, payload)

				req.Header.Add("Content-Type", "application/json")
				req.Header.Add("Authorization", "Basic "+authEncoded)
				res, err := http.DefaultClient.Do(req)
				if res.StatusCode != 200 {
					_ = err
					_ = exec.Command("logger", "-t", "HOSTER_HA_REST", "ERROR: VM START FAILED FOR THE VM: "+v.VmName+" ON: "+v.CurrentHost).Run()
					continue
				}
			}
		}
	}
}

func modifyHostsDb(input ModifyHostsDbStruct) {
	defer func() {
		if r := recover(); r != nil {
			errorValue := fmt.Sprintf("%s", r)
			_ = exec.Command("logger", "-t", "HOSTER_HA_REST", "PANIC AVOIDED: modifyHostsDb(): "+errorValue).Run()
		}
	}()

	// hostsDbLock.Lock()

	if input.addOrUpdate {
		hostFound := false
		hostIndex := 0

		for i, v := range haHostsDb {
			if input.data.NodeInfo.Hostname == v.NodeInfo.Hostname {
				hostFound = true
				hostIndex = i
			}
		}

		if !hostFound {
			_ = exec.Command("logger", "-t", "HOSTER_HA_REST", "INFO: registered a new node: "+input.data.NodeInfo.Hostname).Run()
			haHostsDb = append(haHostsDb, input.data)
		} else {
			// _ = exec.Command("logger", "-t", "HOSTER_HA_REST", "DEBUG: Updated last ping time and network address for "+msg.NodeInfo.Hostname).Run()
			haHostsDb[hostIndex].NodeInfo.Address = input.data.NodeInfo.Address
			if input.data.NodeInfo.StartupTime > 0 {
				haHostsDb[hostIndex].NodeInfo.StartupTime = input.data.NodeInfo.StartupTime
			}
			haHostsDb[hostIndex].LastPing = time.Now().Unix()
		}
	}

	if input.remove {
		for i, v := range haHostsDb {
			if input.data.NodeInfo.Hostname == v.NodeInfo.Hostname && len(v.NodeInfo.Hostname) > 0 {
				haHostsDb[i] = haHostsDb[len(haHostsDb)-1]
				haHostsDb[len(haHostsDb)-1] = HosterHaNodeStruct{}
				haHostsDb = haHostsDb[0 : len(haHostsDb)-1]
				_ = exec.Command("logger", "-t", "HOSTER_HA_REST", "WARN: host has been removed from the cluster: "+input.data.NodeInfo.Hostname).Run()
			}
		}
	}

	// hostsDbLock.Unlock()
}

func readHostsDb() (db []HosterHaNodeStruct) {
	defer func() {
		if r := recover(); r != nil {
			errorValue := fmt.Sprintf("%s", r)
			_ = exec.Command("logger", "-t", "HOSTER_HA_REST", "PANIC AVOIDED: readHostsDb(): "+errorValue).Run()
		}
	}()

	// hostsDbLock.RLock()
	db = haHostsDb
	// hostsDbLock.RUnlock()

	return
}

func terminateOtherMembers() {
	candidateFound := false
	for _, v := range haConfig.Candidates {
		if v.Hostname == myHostname {
			candidateFound = true
		}
	}
	if !candidateFound {
		_ = exec.Command("logger", "-t", "HOSTER_HA_REST", "ERROR: not a candidate node, use one of the candidates to shutdown the whole cluster").Run()
		return
	}

	var wg sync.WaitGroup
	for _, v := range readHostsDb() {
		if v.NodeInfo.Hostname == myHostname {
			continue
		}
		wg.Add(1)
		go func(node HosterHaNodeStruct) {
			_ = exec.Command("logger", "-t", "HOSTER_HA_REST", "INFO: sending a shutdown signal to: "+node.NodeInfo.Hostname).Run()
			url := node.NodeInfo.Protocol + "://" + node.NodeInfo.Address + ":" + node.NodeInfo.Port + "/api/v1/ha/terminate"

			req, err := http.NewRequest("POST", url, nil)
			if err != nil {
				_ = exec.Command("logger", "-t", "HOSTER_HA_REST", "ERROR: could not form the /terminate request: "+err.Error()).Run()
				return
			}

			auth := node.NodeInfo.User + ":" + node.NodeInfo.Password
			authEncoded := base64.StdEncoding.EncodeToString([]byte(auth))
			req.Header.Add("Authorization", "Basic "+authEncoded)
			res, err := http.DefaultClient.Do(req)

			if err != nil {
				_ = exec.Command("logger", "-t", "HOSTER_HA_REST", "WARN: could not notify the member: "+node.NodeInfo.Hostname+". Error: "+err.Error()).Run()
			} else {
				req.Body.Close()
				res.Body.Close()
			}
			wg.Done()
		}(v)
	}
	wg.Wait()
}
