package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"hoster/cmd"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"sort"
	"strconv"
	"strings"
	"time"
)

type NodeStruct struct {
	Hostname         string `json:"hostname"`
	Protocol         string `json:"protocol"`
	Address          string `json:"address"`
	Port             string `json:"port"`
	User             string `json:"user"`
	Password         string `json:"password"`
	FailOverStrategy string `json:"failover_strategy"`
	FailOverTime     int64  `json:"failover_time"`
}

type HosterHaNodeStruct struct {
	IsManager   bool       `json:"is_manager"`
	IsCandidate bool       `json:"is_candidate"`
	IsWorker    bool       `json:"is_worker"`
	LastPing    int64      `json:"last_ping"`
	NodeInfo    NodeStruct `json:"node_info"`
}

type HaConfigJsonStruct struct {
	NodeType         string       `json:"node_type"`
	StartupTime      int64        `json:"startup_time"`
	FailOverStrategy string       `json:"failover_strategy"`
	FailOverTime     int64        `json:"failover_time"`
	Candidates       []NodeStruct `json:"candidates"`
	Manager          NodeStruct   `json:"manager"`
}

var haHostsDb []HosterHaNodeStruct
var haConfig HaConfigJsonStruct

var haChannelAdd = make(chan HosterHaNodeStruct, 100)
var haChannelRemove = make(chan HosterHaNodeStruct, 100)

var iAmManager = false
var iAmCandidate = false
var iAmRegistered = false
var initialRegistrationPerformed = false
var lastManagerContact = time.Now().Unix() + 100
var lastCandidate0Contact = time.Now().Unix() + 100
var lastCandidate1Contact = time.Now().Unix() + 100

var haMode bool
var debugMode bool

func init() {
	haModeEnv := os.Getenv("REST_API_HA_MODE")
	if len(haModeEnv) > 0 {
		haMode = true
	} else {
		_ = exec.Command("logger", "-t", "HOSTER_REST", "STARING REST API SERVER IN REGULAR (NON-HA) MODE").Run()
		return
	}

	debugModeEnv := os.Getenv("REST_API_HA_DEBUG")
	if len(debugModeEnv) > 0 {
		debugMode = true
	}

	go addHaNode(haChannelAdd)
	go removeHaNode(haChannelRemove)

	file, _ := os.ReadFile("/opt/hoster-core/config_files/ha_config.json")
	_ = json.Unmarshal(file, &haConfig)

	if haConfig.FailOverTime < 1 {
		haConfig.FailOverTime = 60
	}
	_ = exec.Command("logger", "-t", "HOSTER_HA_REST", "Cluster failover time is: "+strconv.Itoa(int(haConfig.FailOverTime))+" seconds").Run()

	haConfig.StartupTime = time.Now().UnixMicro()
	if haConfig.NodeType == "candidate" {
		iAmCandidate = true
		iAmManager = false
		go manageOfflineNodes()

		go joinHaCluster()
		go iAmCandidateOnline()
		go managerTemporaryFailover()
	} else if haConfig.NodeType == "manager" {
		iAmManager = true
		iAmCandidate = false

		initializeHaCluster()
		go manageOfflineNodes()
		go iAmManagerOnline()
	} else {
		go joinHaCluster()
		go iAmWorkerOnline()
	}
	go pingPong()
}

func initializeHaCluster() {
	hosterNode := HosterHaNodeStruct{}
	hosterNode.IsCandidate = false
	hosterNode.IsWorker = false
	hosterNode.IsManager = true
	hosterNode.LastPing = time.Now().Unix()
	hosterNode.NodeInfo = haConfig.Manager
	hosterNode.NodeInfo.FailOverStrategy = haConfig.FailOverStrategy
	hosterNode.NodeInfo.FailOverTime = haConfig.FailOverTime

	haChannelAdd <- hosterNode
	iAmRegistered = true
}

func joinHaCluster() {
	for {
		if iAmRegistered {
			time.Sleep(time.Second * 5)
			continue
		}
		user := "admin"
		password := "123456"
		port := 3000

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

		host := NodeStruct{}
		host.Hostname = cmd.GetHostName()
		host.FailOverStrategy = "cireset"
		host.User = user
		host.Password = password
		host.Port = strconv.Itoa(port)
		host.Protocol = "http"
		host.FailOverStrategy = haConfig.FailOverStrategy
		host.FailOverTime = haConfig.FailOverTime

		jsonPayload, _ := json.Marshal(host)
		payload := strings.NewReader(string(jsonPayload))

		url := haConfig.Manager.Protocol + "://" + haConfig.Manager.Address + ":" + haConfig.Manager.Port + "/api/v1/ha/register"
		req, _ := http.NewRequest("POST", url, payload)
		auth := haConfig.Manager.User + ":" + haConfig.Manager.Password
		authEncoded := base64.StdEncoding.EncodeToString([]byte(auth))
		req.Header.Add("Content-Type", "application/json")
		req.Header.Add("Authorization", "Basic "+authEncoded)

		for {
			res, err := http.DefaultClient.Do(req)
			if err != nil {
				_ = exec.Command("logger", "-t", "HOSTER_HA_REST", "Could not join the manager: "+err.Error()).Run()
				time.Sleep(time.Second * 30)
				continue
			}

			defer res.Body.Close()
			body, _ := io.ReadAll(res.Body)

			if !initialRegistrationPerformed {
				_ = exec.Command("logger", "-t", "HOSTER_HA_REST", "Successfully joined the cluster: "+string(body)).Run()
			} else {
				_ = exec.Command("logger", "-t", "HOSTER_HA_REST", "Successfully restored the connection to the cluster manager: "+string(body)).Run()
			}

			iAmRegistered = true
			initialRegistrationPerformed = true
			lastManagerContact = time.Now().Unix()

			break
		}
	}
}

func pingPong() {
	for {
		if iAmRegistered {
			host := NodeStruct{}
			host.Hostname = cmd.GetHostName()

			jsonPayload, _ := json.Marshal(host)
			payload := strings.NewReader(string(jsonPayload))

			if !iAmManager {
				url := haConfig.Manager.Protocol + "://" + haConfig.Manager.Address + ":" + haConfig.Manager.Port + "/api/v1/ha/ping"
				req, _ := http.NewRequest("POST", url, payload)
				auth := haConfig.Manager.User + ":" + haConfig.Manager.Password
				authEncoded := base64.StdEncoding.EncodeToString([]byte(auth))
				req.Header.Add("Content-Type", "application/json")
				req.Header.Add("Authorization", "Basic "+authEncoded)
				_, err := http.DefaultClient.Do(req)
				if err != nil {
					_ = exec.Command("logger", "-t", "HOSTER_HA_REST", "Failed to ping the manager: "+err.Error()).Run()
					iAmRegistered = false
					time.Sleep(time.Second * 10)
					continue
				}
				lastManagerContact = time.Now().Unix()
				time.Sleep(time.Second * 4)
				continue
			} else {
				for _, v := range haConfig.Candidates {
					if v.Hostname == cmd.GetHostName() {
						continue
					}
					url := v.Protocol + "://" + v.Address + ":" + v.Port + "/api/v1/ha/ping"
					req, _ := http.NewRequest("POST", url, payload)
					auth := haConfig.Manager.User + ":" + haConfig.Manager.Password
					authEncoded := base64.StdEncoding.EncodeToString([]byte(auth))
					req.Header.Add("Content-Type", "application/json")
					req.Header.Add("Authorization", "Basic "+authEncoded)
					_, err := http.DefaultClient.Do(req)
					if err != nil {
						_ = exec.Command("logger", "-t", "HOSTER_HA_REST", "Failed to ping the candidate node: "+err.Error()).Run()
						time.Sleep(time.Second * 10)
						continue
					}
					if v.Hostname == haConfig.Candidates[0].Hostname {
						lastCandidate0Contact = time.Now().Unix()
					}
					if v.Hostname == haConfig.Candidates[1].Hostname {
						lastCandidate1Contact = time.Now().Unix()
					}
					time.Sleep(time.Second * 4)
					continue
				}
			}
		} else {
			time.Sleep(time.Second * 10)
			continue
		}
	}
}

func addHaNode(haChannelAdd chan HosterHaNodeStruct) {
	defer func() {
		if r := recover(); r != nil {
			errorValue := fmt.Sprintf("%s", r)
			_ = exec.Command("logger", "-t", "HOSTER_HA_REST", "addHaNode() Recovered from panic: "+errorValue).Run()
		}
	}()

	for msg := range haChannelAdd {
		hostFound := false
		hostIndex := 0
		for i, v := range haHostsDb {
			if msg.NodeInfo.Hostname == v.NodeInfo.Hostname {
				hostFound = true
				hostIndex = i
			}
		}
		if !hostFound {
			_ = exec.Command("logger", "-t", "HOSTER_HA_REST", "Registered a new node: "+msg.NodeInfo.Hostname).Run()
			for _, v := range haConfig.Candidates {
				if v.Hostname == msg.NodeInfo.Hostname {
					msg.IsCandidate = true
					msg.IsManager = false
					msg.IsWorker = false
				} else {
					msg.IsCandidate = false
					msg.IsManager = false
					msg.IsWorker = true
				}
			}
			haHostsDb = append(haHostsDb, msg)
		} else {
			// _ = exec.Command("logger", "-t", "HOSTER_HA_REST", "DEBUG: Updated last ping time and network address for "+msg.NodeInfo.Hostname).Run()
			haHostsDb[hostIndex].NodeInfo.Address = msg.NodeInfo.Address
			haHostsDb[hostIndex].LastPing = time.Now().Unix()
		}
	}
}

func removeHaNode(haChannelRemove chan HosterHaNodeStruct) {
	defer func() {
		if r := recover(); r != nil {
			errorValue := fmt.Sprintf("%s", r)
			_ = exec.Command("logger", "-t", "HOSTER_HA_REST", "removeHaNode() Recovered from panic: "+errorValue).Run()
		}
	}()

	for msg := range haChannelRemove {
		for i, v := range haHostsDb {
			if msg.NodeInfo.Hostname == v.NodeInfo.Hostname {
				haHostsDb[i] = haHostsDb[len(haHostsDb)-1]
				haHostsDb[len(haHostsDb)-1] = HosterHaNodeStruct{}
				haHostsDb = haHostsDb[0 : len(haHostsDb)-1]
				_ = exec.Command("logger", "-t", "HOSTER_HA_REST", "Host has been removed from the cluster: "+msg.NodeInfo.Hostname).Run()
				break
			}
		}
	}
}

func manageOfflineNodes() {
	for {
		for i, v := range haHostsDb {
			if (time.Now().Unix() > v.LastPing+v.NodeInfo.FailOverTime) && !v.IsManager {
				_ = exec.Command("logger", "-t", "HOSTER_HA_REST", "Host has gone offline: "+v.NodeInfo.Hostname).Run()
				failoverHostVms(v)
				haChannelRemove <- haHostsDb[i]
			}
		}
		time.Sleep(time.Second * 10)
	}
}

func failoverHostVms(haNode HosterHaNodeStruct) {
	defer func() {
		if r := recover(); r != nil {
			errorValue := fmt.Sprintf("%s", r)
			_ = exec.Command("logger", "-t", "HOSTER_HA_REST", "failoverHostVms() Recovered from panic: "+errorValue).Run()
		}
	}()

	if !iAmManager {
		time.Sleep(time.Second * 10)
		return
	}

	haVms := []HaVm{}
	for _, v := range haHostsDb {
		haVmsTemp := []HaVm{}
		if v.NodeInfo.Hostname == haNode.NodeInfo.Hostname {
			continue
		}

		url := v.NodeInfo.Protocol + "://" + v.NodeInfo.Address + ":" + v.NodeInfo.Port + "/api/v1/ha/vms"
		req, _ := http.NewRequest("GET", url, nil)
		auth := v.NodeInfo.User + ":" + v.NodeInfo.Password
		authEncoded := base64.StdEncoding.EncodeToString([]byte(auth))
		req.Header.Add("Authorization", "Basic "+authEncoded)
		res, err := http.DefaultClient.Do(req)
		if err != nil {
			_ = exec.Command("logger", "-t", "HOSTER_HA_REST", "Line 309: "+err.Error()).Run()
			continue
		}

		defer res.Body.Close()
		body, _ := io.ReadAll(res.Body)

		err = json.Unmarshal(body, &haVmsTemp)
		if err != nil {
			_ = exec.Command("logger", "-t", "HOSTER_HA_REST", "Line 318: "+err.Error()).Run()
			continue
		}

		for _, vv := range haVmsTemp {
			if vv.ParentHost == haNode.NodeInfo.Hostname {
				haVms = append(haVms, vv)
			}
		}
	}

	sortBySnapshotDate := func(i, j int) bool {
		return haVms[i].LatestSnapshot < haVms[j].LatestSnapshot
	}
	sort.Slice(haVms, sortBySnapshotDate)

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
			_ = exec.Command("logger", "-t", "HOSTER_HA_REST", "DEBUG: FAILING OVER VM: "+v.VmName+" FROM: "+v.ParentHost+" TO: "+v.CurrentHost).Run()
			continue
		}
		for _, vv := range haHostsDb {
			if vv.NodeInfo.Hostname == v.CurrentHost {
				_ = exec.Command("logger", "-t", "HOSTER_HA_REST", "FAILING OVER VM: "+v.VmName+" FROM: "+v.ParentHost+" TO: "+v.CurrentHost).Run()

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
						_ = exec.Command("logger", "-t", "HOSTER_HA_REST", "CIRESET FAILED FOR THE VM: "+v.VmName+" ON: "+v.CurrentHost).Run()
						continue
					}
				} else if vv.NodeInfo.FailOverStrategy == "changeparent" || vv.NodeInfo.FailOverStrategy == "change-parent" {
					url := vv.NodeInfo.Protocol + "://" + vv.NodeInfo.Address + ":" + vv.NodeInfo.Port + "/api/v1/vm/change-parent"
					payload := strings.NewReader(`{ "name": "` + v.VmName + `" }`)
					req, _ := http.NewRequest("POST", url, payload)

					req.Header.Add("Content-Type", "application/json")
					req.Header.Add("Authorization", "Basic "+authEncoded)
					res, err := http.DefaultClient.Do(req)
					if res.StatusCode != 200 {
						_ = err
						_ = exec.Command("logger", "-t", "HOSTER_HA_REST", "CHANGE PARENT FAILED FOR THE VM: "+v.VmName+" ON: "+v.CurrentHost).Run()
						continue
					}
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
					_ = exec.Command("logger", "-t", "HOSTER_HA_REST", "VM START FAILED FOR THE VM: "+v.VmName+" ON: "+v.CurrentHost).Run()
					continue
				}
			}
		}
	}
}

func iAmWorkerOnline() {
	for {
		managerOffline := time.Now().Unix() > lastManagerContact+haConfig.FailOverTime
		candidateZeroOffline := time.Now().Unix() > lastCandidate0Contact+haConfig.FailOverTime
		candidateOneOffline := time.Now().Unix() > lastCandidate1Contact+haConfig.FailOverTime
		if initialRegistrationPerformed {
			if (managerOffline && candidateZeroOffline) || (managerOffline && candidateOneOffline) || (candidateZeroOffline && candidateOneOffline) {
				_ = exec.Command("logger", "-t", "HOSTER_HA_REST", "Could not reach cluster manager and one of the candidates for "+strconv.Itoa(int(haConfig.FailOverTime))+" seconds, exiting the process").Run()
				os.Exit(1)
			}
		}
		time.Sleep(time.Second * 5)
	}
}

func iAmManagerOnline() {
	for iAmManager {
		candidateZeroOffline := time.Now().Unix() > lastCandidate0Contact+haConfig.FailOverTime
		candidateOneOffline := time.Now().Unix() > lastCandidate1Contact+haConfig.FailOverTime
		if candidateZeroOffline && candidateOneOffline {
			_ = exec.Command("logger", "-t", "HOSTER_HA_REST", "Could not reach cluster candidates for "+strconv.Itoa(int(haConfig.FailOverTime))+" seconds, exiting the process").Run()
			os.Exit(1)
		}
		time.Sleep(time.Second * 5)
	}
}

func iAmCandidateOnline() {
	for {
		if iAmCandidate {
			var otherCandidateOffline bool
			if cmd.GetHostName() == haConfig.Candidates[0].Hostname {
				otherCandidateOffline = time.Now().Unix() > lastCandidate1Contact+haConfig.FailOverTime
			}
			if cmd.GetHostName() == haConfig.Candidates[1].Hostname {
				otherCandidateOffline = time.Now().Unix() > lastCandidate0Contact+haConfig.FailOverTime
			}
			managerOffline := time.Now().Unix() > lastManagerContact+haConfig.FailOverTime

			if otherCandidateOffline && managerOffline {
				_ = exec.Command("logger", "-t", "HOSTER_HA_REST", "Could not reach other cluster candidates for "+strconv.Itoa(int(haConfig.FailOverTime))+" seconds, exiting the process").Run()
				os.Exit(1)
			}
			time.Sleep(time.Second * 5)
		}
	}
}

func managerTemporaryFailover() {
	for {
		if iAmCandidate {
			var otherCandidateOffline bool
			if cmd.GetHostName() == haConfig.Candidates[0].Hostname {
				otherCandidateOffline = time.Now().Unix() > lastCandidate1Contact+haConfig.FailOverTime
			}
			if cmd.GetHostName() == haConfig.Candidates[1].Hostname {
				otherCandidateOffline = time.Now().Unix() > lastCandidate0Contact+haConfig.FailOverTime
			}
			managerOffline := time.Now().Unix() > lastManagerContact+haConfig.FailOverTime

			if !otherCandidateOffline && managerOffline {
				_ = exec.Command("logger", "-t", "HOSTER_HA_REST", "Could not reach our manager for "+strconv.Itoa(int(haConfig.FailOverTime))+", I am the manager now").Run()
				iAmManager = true
			} else {
				iAmManager = false
			}

			time.Sleep(time.Minute * 1)
		}
	}
}
