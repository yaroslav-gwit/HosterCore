package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"hoster/cmd"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
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
}

type HosterHaNodeStruct struct {
	IsManager bool           `json:"is_manager"`
	LastPing  int64          `json:"last_ping"`
	NodeInfo  NodeInfoStruct `json:"node_info"`
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

var haChannelAdd = make(chan HosterHaNodeStruct, 100)
var haChannelRemove = make(chan HosterHaNodeStruct, 100)

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

	go addHaNode(haChannelAdd)
	go removeHaNode(haChannelRemove)

	file, _ := os.ReadFile("/opt/hoster-core/config_files/ha_config.json")
	err = json.Unmarshal(file, &haConfig)
	if err != nil {
		_ = exec.Command("logger", "-t", "HOSTER_HA", "PANIC: could not parse the HA configuration file: "+err.Error()).Run()
		panic("Cannot parse the HA configuration file: " + err.Error())
	}

	if haConfig.FailOverTime < 1 {
		haConfig.FailOverTime = 60
	}
	_ = exec.Command("logger", "-t", "HOSTER_HA_REST", "INFO: cluster failover time is: "+strconv.Itoa(int(haConfig.FailOverTime))+" seconds").Run()

	haConfig.StartupTime = time.Now().UnixMilli()
	go registerNode()
	go trackCandidatesOnline()
	go trackManager()
	go sendPing()
}

var candidatesRegistered = 0
var clusterInitialized = false
var iAmManager = false

func trackCandidatesOnline() {
	defer func() {
		if r := recover(); r != nil {
			errorValue := fmt.Sprintf("%s", r)
			_ = exec.Command("logger", "-t", "HOSTER_HA_REST", "AVOIDED PANIC: trackCandidatesOnline(): "+errorValue).Run()
		}
	}()

	for {
		candidatesRegistered = 0
		for _, v := range haConfig.Candidates {
			if v.Registered {
				candidatesRegistered += 1
			} else {
				candidatesRegistered -= 1
			}
		}

		if !clusterInitialized && candidatesRegistered >= 3 {
			clusterInitialized = true
		}

		if clusterInitialized && candidatesRegistered < 2 {
			_ = exec.Command("logger", "-t", "HOSTER_HA_REST", "EMERG: candidatesRegistered has gone below 2, initiating self fencing").Run()
			os.Exit(0)
		}

		time.Sleep(time.Second * 5)
	}
}

func trackManager() {
	defer func() {
		if r := recover(); r != nil {
			errorValue := fmt.Sprintf("%s", r)
			_ = exec.Command("logger", "-t", "HOSTER_HA_REST", "AVOIDED PANIC: trackManager(): "+errorValue).Run()
		}
	}()

	for {
		if clusterInitialized {
			var myNumber int64
			var clusterNumber int64
			var myHostname = cmd.GetHostName()

			for _, v := range haHostsDb {
				for _, vv := range haConfig.Candidates {
					if myHostname == vv.Hostname {
						myNumber = v.NodeInfo.StartupTime
						continue
					}
					if v.NodeInfo.Hostname == vv.Hostname {
						if clusterNumber < v.NodeInfo.StartupTime {
							clusterNumber = v.NodeInfo.StartupTime
						}
						continue
					}
				}
			}
			if myNumber > clusterNumber {
				if !iAmManager {
					_ = exec.Command("logger", "-t", "HOSTER_HA_REST", "INFO: becoming new cluster manager").Run()
				}
				iAmManager = true
			} else {
				if iAmManager {
					_ = exec.Command("logger", "-t", "HOSTER_HA_REST", "INFO: stepping down as a cluster manager").Run()
				}
				iAmManager = false
			}
		}
		time.Sleep(time.Second * 20)
	}
}

func registerNode() {
	defer func() {
		if r := recover(); r != nil {
			errorValue := fmt.Sprintf("%s", r)
			_ = exec.Command("logger", "-t", "HOSTER_HA_REST", "AVOIDED PANIC: registerNode(): "+errorValue).Run()
		}
	}()

	for {
		if candidatesRegistered >= 3 {
			time.Sleep(time.Second * 10)
			continue
		}
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

			jsonPayload, _ := json.Marshal(host)
			payload := strings.NewReader(string(jsonPayload))

			url := v.Protocol + "://" + v.Address + ":" + v.Port + "/api/v1/ha/register"
			req, _ := http.NewRequest("POST", url, payload)
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
			}
			_ = res
		}
		time.Sleep(time.Second * 10)
	}
}

func sendPing() {
	defer func() {
		if r := recover(); r != nil {
			errorValue := fmt.Sprintf("%s", r)
			_ = exec.Command("logger", "-t", "HOSTER_HA_REST", "AVOIDED PANIC: registerNode(): "+errorValue).Run()
		}
	}()

	for {
		if !clusterInitialized {
			time.Sleep(time.Second * 10)
			continue
		}
		for i, v := range haConfig.Candidates {
			if !v.Registered {
				continue
			}
			go func(i int, v NodeInfoStruct) {
				host := NodeInfoStruct{}
				host.Hostname = cmd.GetHostName()

				jsonPayload, _ := json.Marshal(host)
				payload := strings.NewReader(string(jsonPayload))

				url := v.Protocol + "://" + v.Address + ":" + v.Port + "/api/v1/ha/ping"
				req, _ := http.NewRequest("POST", url, payload)
				auth := v.User + ":" + v.Password
				authEncoded := base64.StdEncoding.EncodeToString([]byte(auth))
				req.Header.Add("Content-Type", "application/json")
				req.Header.Add("Authorization", "Basic "+authEncoded)
				_, err := http.DefaultClient.Do(req)
				if err != nil {
					_ = exec.Command("logger", "-t", "HOSTER_HA_REST", "WARN: failed to ping the candidate node: "+err.Error()).Run()
					haConfig.Candidates[i].Registered = false
				}
			}(i, v)
		}
		time.Sleep(time.Second * 5)
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
			_ = exec.Command("logger", "-t", "HOSTER_HA_REST", "INFO: registered a new node: "+msg.NodeInfo.Hostname).Run()
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
			_ = exec.Command("logger", "-t", "HOSTER_HA_REST", "AVOIDED PANIC: removeHaNode(): "+errorValue).Run()
		}
	}()

	for msg := range haChannelRemove {
		for i, v := range haHostsDb {
			if msg.NodeInfo.Hostname == v.NodeInfo.Hostname {
				haHostsDb[i] = haHostsDb[len(haHostsDb)-1]
				haHostsDb[len(haHostsDb)-1] = HosterHaNodeStruct{}
				haHostsDb = haHostsDb[0 : len(haHostsDb)-1]
				_ = exec.Command("logger", "-t", "HOSTER_HA_REST", "WARN: host has been removed from the cluster: "+msg.NodeInfo.Hostname).Run()
				break
			}
		}
	}
}
