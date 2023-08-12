package main

import (
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
	"time"

	"hoster/cmd"

	"github.com/gofiber/fiber/v2"
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
var lastManagerContact = time.Now().Unix()

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

	_ = iAmCandidate
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
	} else if haConfig.NodeType == "manager" {
		iAmManager = true
		iAmCandidate = false
		initializeHaCluster()
		go manageOfflineNodes()
	} else {
		go joinHaCluster()
		go pingPong()
		go ensureManagerConnection()
	}
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
			haHostsDb = append(haHostsDb, msg)
		} else {
			// _ = exec.Command("logger", "-t", "HOSTER_HA_REST", "Updated last ping time: "+msg.NodeInfo.Hostname).Run()
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
				_ = exec.Command("logger", "-t", "HOSTER_HA_REST", "Failed node has been removed from the cluster: "+msg.NodeInfo.Hostname).Run()
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
				// payloadMap := make(map[string]string)
				// payloadMap["name"] = v.VmName
				// jsonPayload, _ := json.Marshal(payloadMap)
				payload := strings.NewReader(`{ "name": "` + v.VmName + `" }`)
				auth := vv.NodeInfo.User + ":" + vv.NodeInfo.Password
				authEncoded := base64.StdEncoding.EncodeToString([]byte(auth))

				// Use failover strategy to failover the VM
				if vv.NodeInfo.FailOverStrategy == "cireset" || vv.NodeInfo.FailOverStrategy == "ci-reset" {
					url := vv.NodeInfo.Protocol + "://" + vv.NodeInfo.Address + ":" + vv.NodeInfo.Port + "/api/v1/vm/cireset"
					req, _ := http.NewRequest("POST", url, payload)

					req.Header.Add("Content-Type", "application/json")
					req.Header.Add("Authorization", "Basic "+authEncoded)
					res, err := http.DefaultClient.Do(req)
					if res.StatusCode != 200 {
						_ = err
						_ = exec.Command("logger", "-t", "HOSTER_HA_REST", "FAILED TO MOVE THE VM: "+v.VmName+" FROM: "+v.ParentHost+" TO: "+v.CurrentHost+"; ERROR: ").Run()
						continue
					}
				} else if vv.NodeInfo.FailOverStrategy == "changeparent" || vv.NodeInfo.FailOverStrategy == "change-parent" {
					url := vv.NodeInfo.Protocol + "://" + vv.NodeInfo.Address + ":" + vv.NodeInfo.Port + "/api/v1/vm/change-parent"
					req, _ := http.NewRequest("POST", url, payload)

					req.Header.Add("Content-Type", "application/json")
					req.Header.Add("Authorization", "Basic "+authEncoded)
					res, err := http.DefaultClient.Do(req)
					if res.StatusCode != 200 {
						_ = err
						_ = exec.Command("logger", "-t", "HOSTER_HA_REST", "FAILED TO MOVE THE VM: "+v.VmName+" FROM: "+v.ParentHost+" TO: "+v.CurrentHost+"; ERROR: ").Run()
						continue
					}
				}

				// Start VM on a new host
				url := vv.NodeInfo.Protocol + "://" + vv.NodeInfo.Address + ":" + vv.NodeInfo.Port + "/api/v1/vm/start"
				req, _ := http.NewRequest("POST", url, payload)

				req.Header.Add("Content-Type", "application/json")
				req.Header.Add("Authorization", "Basic "+authEncoded)
				res, err := http.DefaultClient.Do(req)
				if res.StatusCode == 200 {
					_ = err
					_ = exec.Command("logger", "-t", "HOSTER_HA_REST", "FAILING OVER VM: "+v.VmName+" FROM: "+v.ParentHost+" TO: "+v.CurrentHost).Run()
					continue
				} else {
					_ = exec.Command("logger", "-t", "HOSTER_HA_REST", "FAILED TO MOVE THE VM: "+v.VmName+" FROM: "+v.ParentHost+" TO: "+v.CurrentHost+"; ERROR: ").Run()
					continue
				}
			}
		}
	}
}

func ensureManagerConnection() {
	for {
		if (time.Now().Unix() > lastManagerContact+haConfig.FailOverTime) && initialRegistrationPerformed {
			_ = exec.Command("logger", "-t", "HOSTER_HA_REST", "Could not reach cluster manager for "+strconv.Itoa(int(haConfig.FailOverTime))+" seconds, exiting the process").Run()
			os.Exit(1)
		}
		time.Sleep(time.Second * 5)
	}
}

func handleHaManagerRegistration(fiberContext *fiber.Ctx) error {
	tagCustomError = ""

	hosterNode := new(NodeStruct)
	if err := fiberContext.BodyParser(hosterNode); err != nil {
		tagCustomError = err.Error()
		fiberContext.Status(fiber.StatusUnprocessableEntity)
		return fiberContext.JSON(fiber.Map{"error": err.Error()})
	}

	hosterHaNode := HosterHaNodeStruct{}
	hosterHaNode.IsManager = false
	hosterHaNode.IsCandidate = false
	hosterHaNode.IsWorker = true
	hosterHaNode.NodeInfo = *hosterNode
	hosterHaNode.NodeInfo.Address = fiberContext.IP()
	hosterHaNode.LastPing = time.Now().Unix()

	haChannelAdd <- hosterHaNode
	return fiberContext.JSON(fiber.Map{"message": "done", "context": hosterHaNode})
}

func handleHaPing(fiberContext *fiber.Ctx) error {
	tagCustomError = ""
	hosterNode := new(NodeStruct)
	if err := fiberContext.BodyParser(hosterNode); err != nil {
		tagCustomError = err.Error()
		fiberContext.Status(fiber.StatusUnprocessableEntity)
		return fiberContext.JSON(fiber.Map{"error": err.Error()})
	}

	hosterHaNode := HosterHaNodeStruct{}
	hosterHaNode.NodeInfo = *hosterNode
	hosterHaNode.NodeInfo.Address = fiberContext.IP()
	hosterHaNode.LastPing = time.Now().Unix()

	haChannelAdd <- hosterHaNode
	return fiberContext.JSON(fiber.Map{"message": "pong"})
}

func handleHaTerminate(fiberContext *fiber.Ctx) error {
	return fiberContext.JSON(fiber.Map{"message": "done"})
}

func handleHaPromote(fiberContext *fiber.Ctx) error {
	return fiberContext.JSON(fiber.Map{"message": "done"})
}

func handleHaMonitor(fiberContext *fiber.Ctx) error {
	return fiberContext.JSON(fiber.Map{"message": "done"})
}

func handleHaShareWorkers(fiberContext *fiber.Ctx) error {
	workers := []HosterHaNodeStruct{}
	for _, v := range haHostsDb {
		if v.IsWorker {
			workers = append(workers, v)
		}
	}
	return fiberContext.JSON(workers)
}

func handleHaShareManager(fiberContext *fiber.Ctx) error {
	manager := HosterHaNodeStruct{}
	for _, v := range haHostsDb {
		if v.IsManager {
			manager = v
		}
	}
	return fiberContext.JSON(manager)
}

func handleHaShareCandidates(fiberContext *fiber.Ctx) error {
	candidates := []HosterHaNodeStruct{}
	for _, v := range haHostsDb {
		if v.IsCandidate {
			candidates = append(candidates, v)
		}
	}
	return fiberContext.JSON(candidates)
}

func handleHaShareAllMembers(fiberContext *fiber.Ctx) error {
	if iAmManager {
		return fiberContext.JSON(haHostsDb)
	} else {
		return fiberContext.JSON([]HosterHaNodeStruct{})
	}
}

type HaVm struct {
	VmName         string `json:"vm_name"`
	Live           bool   `json:"live"`
	LatestSnapshot string `json:"latest_snapshot"`
	ParentHost     string `json:"parent_host"`
	CurrentHost    string `json:"current_host"`
}

func haVmsList() []HaVm {
	defer func() {
		if r := recover(); r != nil {
			errorValue := fmt.Sprintf("%s", r)
			_ = exec.Command("logger", "-t", "HOSTER_HA_REST", "haVmsList() Recovered from error: "+errorValue).Run()
		}
	}()

	haVms := []HaVm{}
	vmList := cmd.GetAllVms()

	for _, vm := range vmList {
		vmFound := false
		for _, haVm := range haVms {
			if vm == haVm.VmName {
				vmFound = true
			}
		}
		if vmFound {
			continue
		}

		vmConfig := cmd.VmConfig(vm)
		if !cmd.VmIsInProduction(vmConfig.LiveStatus) {
			continue
		}

		haVmTemp := HaVm{}
		haVmTemp.VmName = vm
		haVmTemp.ParentHost = vmConfig.ParentHost
		haVmTemp.CurrentHost = cmd.GetHostName()
		haVmTemp.Live = cmd.VmLiveCheck(vm)
		snapshotList, _ := cmd.GetSnapshotInfo(vm, true)
		if len(snapshotList) > 0 {
			haVmTemp.LatestSnapshot = snapshotList[len(snapshotList)-1].Name
		}

		haVms = append(haVms, haVmTemp)
	}

	return haVms
}

func handleHaVmsList(fiberContext *fiber.Ctx) error {
	return fiberContext.JSON(haVmsList())
}
