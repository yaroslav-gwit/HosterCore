package HandlersHA

import (
	RestApiConfig "HosterCore/internal/app/rest_api_v2/pkg/config"
	FreeBSDsysctls "HosterCore/internal/pkg/freebsd/sysctls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

var candidatesRegistered = 0
var clusterInitialized = false
var iAmManager = false
var myHostname, _ = FreeBSDsysctls.SysctlKernHostname()

func trackCandidatesOnline() {
	defer func() {
		if r := recover(); r != nil {
			errorValue := fmt.Sprintf("%s", r)
			// _ = exec.Command("logger", "-t", "HOSTER_HA_REST", "PANIC AVOIDED: trackCandidatesOnline(): "+errorValue).Run()
			internalLog.Warnf("PANIC AVOIDED: trackCandidatesOnline() %s", errorValue)
		}
	}()

	for {
		candidatesRegistered = 0

		for _, v := range haConf.Candidates {
			if v.Registered {
				candidatesRegistered += 1
			}
		}

		if !clusterInitialized && candidatesRegistered >= 3 {
			clusterInitialized = true
		}

		if clusterInitialized && candidatesRegistered < 2 {
			// _ = exec.Command("logger", "-t", "HOSTER_HA_REST", "EMERG: candidatesRegistered has gone below 2, initiating self fencing").Run()
			internalLog.Warn("number of manager nodes (candidatesRegistered) has gone below 2, initiating self fencing")
			os.Exit(0)
		}

		time.Sleep(time.Second * 10)
	}
}

func trackManager() {
	defer func() {
		if r := recover(); r != nil {
			errorValue := fmt.Sprintf("%s", r)
			// _ = exec.Command("logger", "-t", "HOSTER_HA_REST", "PANIC AVOIDED: trackManager(): "+errorValue).Run()
			internalLog.Warnf("PANIC AVOIDED: trackManager() %s", errorValue)
		}
	}()

	for {
		if clusterInitialized {
			var copyHostsDb = readHostsDb(&hostsDbLock)
			if len(copyHostsDb) < 2 {
				continue
			}

			var filteredCandidates []HosterHaNode

			sort.Slice(copyHostsDb, func(i int, j int) bool {
				return copyHostsDb[i].NodeInfo.StartupTime < copyHostsDb[j].NodeInfo.StartupTime
			})

			for _, host := range copyHostsDb {
				for _, candidate := range haConf.Candidates {
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
					// _ = exec.Command("logger", "-t", "HOSTER_HA_REST", "INFO: becoming a new cluster manager").Run()
					internalLog.Info("becoming a new cluster manager")
					iAmManager = true
				}
			} else {
				if iAmManager {
					// _ = exec.Command("logger", "-t", "HOSTER_HA_REST", "WARN: stepping down as a cluster manager").Run()
					internalLog.Warn("stepping down as a cluster manager")
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
			// _ = exec.Command("logger", "-t", "HOSTER_HA_REST", "PANIC AVOIDED: registerNode(): "+errorValue).Run()
			internalLog.Warnf("PANIC AVOIDED: registerNode() %s", errorValue)
		}
	}()

	for {
		if candidatesRegistered >= 3 {
			time.Sleep(time.Second * 10)
			continue
		}

		time.Sleep(time.Second * 10)
		for i, v := range haConf.Candidates {
			if v.Registered {
				continue
			}
			host := RestApiConfig.HaNode{}
			host.Hostname, _ = FreeBSDsysctls.SysctlKernHostname()
			host.FailOverStrategy = haConf.FailOverStrategy

			configFound := false
			for _, vv := range restConf.HTTPAuth {
				if vv.HaUser {
					host.User = vv.User
					host.Password = vv.Password
					host.Port = strconv.Itoa(restConf.Port)
					host.Protocol = restConf.Protocol
					configFound = true
					break
				}
			}
			if !configFound {
				// _ = exec.Command("logger", "-t", "HOSTER_HA_REST", "PANIC: registerNode(): could not configure the API server correctly").Run()
				internalLog.Warn("registerNode(): could not configure the API server correctly")
				os.Exit(1)
			}

			host.FailOverStrategy = haConf.FailOverStrategy
			host.FailOverTime = haConf.FailOverTime
			host.StartupTime = haConf.StartupTime
			host.BackupNode = haConf.BackupNode

			jsonPayload, _ := json.Marshal(host)
			payload := strings.NewReader(string(jsonPayload))
			url := v.Protocol + "://" + v.Address + ":" + v.Port + "/api/v2/ha/register"

			req, err := http.NewRequest("POST", url, payload)
			if err != nil {
				// _ = exec.Command("logger", "-t", "HOSTER_HA_REST", "ERROR: could not form the /register request: "+err.Error()).Run()
				internalLog.Error("could not form the /register request: " + err.Error())
				time.Sleep(time.Second * 10)
				continue
			}

			auth := v.User + ":" + v.Password
			authEncoded := base64.StdEncoding.EncodeToString([]byte(auth))
			req.Header.Add("Content-Type", "application/json")
			req.Header.Add("Authorization", "Basic "+authEncoded)

			res, err := http.DefaultClient.Do(req)
			if err != nil {
				// _ = exec.Command("logger", "-t", "HOSTER_HA_REST", "WARN: could not join the candidate: "+err.Error()).Run()
				internalLog.Error("could not join the other candidate: " + err.Error())
				time.Sleep(time.Second * 30)
				continue
			} else {
				// _ = exec.Command("logger", "-t", "HOSTER_HA_REST", "SUCCESS: joined the candidate: "+v.Hostname).Run()
				internalLog.Info("joined the candidate " + v.Hostname)
				haConf.Candidates[i].Registered = true
				haConf.Candidates[i].StartupTime = haConf.StartupTime
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
			// _ = exec.Command("logger", "-t", "HOSTER_HA_REST", "PANIC AVOIDED: sendPing(): "+errorValue).Run()
			internalLog.Warnf("PANIC AVOIDED: sendPing() %s", errorValue)
		}
	}()

	for {
		if !clusterInitialized {
			time.Sleep(time.Second * 10)
			continue
		}

		var wg sync.WaitGroup
		wg.Add(len(haConf.Candidates))

		for i, v := range haConf.Candidates {
			if !v.Registered {
				continue
			}

			go func(i int, v RestApiConfig.HaNode) {
				defer func() {
					r := recover()
					if r != nil {
						errorValue := fmt.Sprintf("%s", r)
						// _ = exec.Command("logger", "-t", "HOSTER_HA_REST", "PANIC AVOIDED: sendPingMainGoRoutine(): "+errorValue).Run()
						internalLog.Warnf("PANIC AVOIDED: sendPingMainGoRoutine() %s", errorValue)
					}
					wg.Done()
				}()

				host := RestApiConfig.HaNode{}
				host.Hostname, _ = FreeBSDsysctls.SysctlKernHostname()
				host.StartupTime = haConf.StartupTime

				jsonPayload, err := json.Marshal(host)
				if err != nil {
					// _ = exec.Command("logger", "-t", "HOSTER_HA_REST", "ERR: pingGoRoutineJSONPAYLOAD(): "+err.Error()).Run()
					internalLog.Error("could not form a JSON payload pingGoRoutine(): " + err.Error())
					return
				}

				payload := strings.NewReader(string(jsonPayload))
				url := v.Protocol + "://" + v.Address + ":" + v.Port + "/api/v2/ha/ping"
				req, err := http.NewRequest("POST", url, payload)
				if err != nil {
					// _ = exec.Command("logger", "-t", "HOSTER_HA_REST", "ERR: pingGoRoutineREQ(): "+err.Error()).Run()
					internalLog.Error("pingGoRoutineRequest(): " + err.Error())
					return
				}

				auth := v.User + ":" + v.Password
				authEncoded := base64.StdEncoding.EncodeToString([]byte(auth))
				req.Header.Add("Content-Type", "application/json")
				req.Header.Add("Authorization", "Basic "+authEncoded)
				resp, err := http.DefaultClient.Do(req)
				if err != nil {
					// _ = exec.Command("logger", "-t", "HOSTER_HA_REST", "WARN: failed to ping the candidate node: "+err.Error()).Run()
					internalLog.Warn("failed to ping the candidate node: " + err.Error())
					haConf.Candidates[i].TimesFailed += 1
					if haConf.Candidates[i].TimesFailed >= 3 {
						haConf.Candidates[i].Registered = false
					}
				} else {
					if haConf.Candidates[i].TimesFailed > 0 {
						haConf.Candidates[i].TimesFailed -= 1
					}
					haConf.Candidates[i].Registered = true
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
			// _ = exec.Command("logger", "-t", "HOSTER_HA_REST", "PANIC AVOIDED: removeOfflineNodes(): "+errorValue).Run()
			internalLog.Warnf("PANIC AVOIDED: removeOfflineNodes() %s", errorValue)
		}
	}()

	for {
		hostsDbCopy := readHostsDb(&hostsDbLock)
		for _, v := range hostsDbCopy {
			if time.Now().Unix() > v.LastPing+v.NodeInfo.FailOverTime {
				if len(v.NodeInfo.Hostname) > 0 {
					failoverHostVms(v)
					modifyHostsDb(ModifyHostsDb{data: v, remove: true}, &hostsDbLock)
					// _ = exec.Command("logger", "-t", "HOSTER_HA_REST", "WARN: host has gone offline: "+v.NodeInfo.Hostname).Run()
					internalLog.Warnf("host has gone offline %s", v.NodeInfo.Hostname)
				}
			}
		}
		time.Sleep(time.Second * 4)
	}
}

func failoverHostVms(haNode HosterHaNode) {
	defer func() {
		if r := recover(); r != nil {
			errorValue := fmt.Sprintf("%s", r)
			// _ = exec.Command("logger", "-t", "HOSTER_HA_REST", "PANIC AVOIDED: failoverHostVms(): "+errorValue).Run()
			internalLog.Warnf("PANIC AVOIDED: failoverHostVms() %s", errorValue)
		}
	}()

	if !iAmManager {
		time.Sleep(time.Second * 3)
		return
	}

	haVms := []HaVm{}
	hostsDbCopy := readHostsDb(&hostsDbLock)
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

		url := v.NodeInfo.Protocol + "://" + v.NodeInfo.Address + ":" + v.NodeInfo.Port + "/api/v2/ha/vms"
		req, _ := http.NewRequest("GET", url, nil)
		auth := v.NodeInfo.User + ":" + v.NodeInfo.Password
		authEncoded := base64.StdEncoding.EncodeToString([]byte(auth))
		req.Header.Add("Authorization", "Basic "+authEncoded)
		res, err := http.DefaultClient.Do(req)
		if err != nil {
			// _ = exec.Command("logger", "-t", "HOSTER_HA_REST", "ERROR: line 345: "+err.Error()).Run()
			internalLog.Error("line 333: " + err.Error())
			continue
		}

		defer res.Body.Close()
		body, _ := io.ReadAll(res.Body)

		err = json.Unmarshal(body, &haVmsTemp)
		if err != nil {
			// _ = exec.Command("logger", "-t", "HOSTER_HA_REST", "ERROR: line 354: "+err.Error()).Run()
			internalLog.Error("line 343: " + err.Error())
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
		if restConf.HaDebug {
			// _ = exec.Command("logger", "-t", "HOSTER_HA_REST", "WARN: MOVING VM: "+v.VmName+" FROM offline parent: "+v.ParentHost+" TO: "+v.CurrentHost).Run()
			internalLog.Warnf("failing over a VM ::%s:: from an offline host ::%s:: to ::%s::", v.VmName, v.ParentHost, v.CurrentHost)
			continue
		}

		for _, vv := range hostsDbCopy {
			if vv.NodeInfo.Hostname == v.CurrentHost {
				time.Sleep(1500 * time.Millisecond)
				// _ = exec.Command("logger", "-t", "HOSTER_HA_REST", "WARN: MOVING VM: "+v.VmName+" FROM offline parent: "+v.ParentHost+" TO: "+v.CurrentHost).Run()
				internalLog.Warnf("failing over a VM ::%s:: from an offline host ::%s:: to ::%s::", v.VmName, v.ParentHost, v.CurrentHost)

				auth := vv.NodeInfo.User + ":" + vv.NodeInfo.Password
				authEncoded := base64.StdEncoding.EncodeToString([]byte(auth))

				// Use failover strategy to failover the VM
				if vv.NodeInfo.FailOverStrategy == "cireset" || vv.NodeInfo.FailOverStrategy == "ci-reset" {
					url := vv.NodeInfo.Protocol + "://" + vv.NodeInfo.Address + ":" + vv.NodeInfo.Port + "/api/v2/vm/cireset"
					payload := strings.NewReader(`{ "name": "` + v.VmName + `" }`)
					req, _ := http.NewRequest("POST", url, payload)

					req.Header.Add("Content-Type", "application/json")
					req.Header.Add("Authorization", "Basic "+authEncoded)
					res, err := http.DefaultClient.Do(req)
					if res.StatusCode != 200 {
						_ = err
						// _ = exec.Command("logger", "-t", "HOSTER_HA_REST", "ERROR: CIRESET FAILED FOR THE VM: "+v.VmName+" ON: "+v.CurrentHost).Run()
						internalLog.Errorf("cireset call failed for the VM ::%s:: on host ::%s::", v.VmName, v.CurrentHost)
						continue
					}
					req.Body.Close()
					res.Body.Close()
				} else if vv.NodeInfo.FailOverStrategy == "changeparent" || vv.NodeInfo.FailOverStrategy == "change-parent" {
					url := vv.NodeInfo.Protocol + "://" + vv.NodeInfo.Address + ":" + vv.NodeInfo.Port + "/api/v2/vm/change-parent"
					payload := strings.NewReader(`{ "name": "` + v.VmName + `" }`)
					req, _ := http.NewRequest("POST", url, payload)

					req.Header.Add("Content-Type", "application/json")
					req.Header.Add("Authorization", "Basic "+authEncoded)
					res, err := http.DefaultClient.Do(req)
					if res.StatusCode != 200 {
						_ = err
						// _ = exec.Command("logger", "-t", "HOSTER_HA_REST", "ERROR: CHANGE PARENT FAILED FOR THE VM: "+v.VmName+" ON: "+v.CurrentHost).Run()
						internalLog.Errorf("change parent call failed for the VM ::%s:: on host ::%s::", v.VmName, v.CurrentHost)
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
					// _ = exec.Command("logger", "-t", "HOSTER_HA_REST", "ERROR: VM START FAILED FOR THE VM: "+v.VmName+" ON: "+v.CurrentHost).Run()
					internalLog.Errorf("start call failed for the VM ::%s:: on host ::%s::", v.VmName, v.CurrentHost)
					continue
				}
			}
		}
	}
}

func modifyHostsDb(input ModifyHostsDb, dbLock *sync.RWMutex) {
	defer func() {
		if r := recover(); r != nil {
			errorValue := fmt.Sprintf("%s", r)
			// _ = exec.Command("logger", "-t", "HOSTER_HA_REST", "PANIC AVOIDED: modifyHostsDb(): "+errorValue).Run()
			internalLog.Warnf("PANIC AVOIDED: modifyHostsDb() %s", errorValue)
		}
		// dbLock.Unlock()
	}()

	defer dbLock.Unlock()
	if input.addOrUpdate {
		hostFound := false
		hostIndex := 0

		dbLock.Lock()
		for i, v := range haHostsDb {
			if input.data.NodeInfo.Hostname == v.NodeInfo.Hostname {
				hostFound = true
				hostIndex = i
			}
		}

		if !hostFound {
			// _ = exec.Command("logger", "-t", "HOSTER_HA_REST", "INFO: registered a new node: "+input.data.NodeInfo.Hostname).Run()
			internalLog.Infof("registered a new node: %s", input.data.NodeInfo.Hostname)
			haHostsDb = append(haHostsDb, input.data)
		} else {
			// _ = exec.Command("logger", "-t", "HOSTER_HA_REST", "DEBUG: Updated last ping time and network address for "+msg.NodeInfo.Hostname).Run()
			internalLog.Debugf("updated last ping time for: %s", input.data.NodeInfo.Hostname)
			haHostsDb[hostIndex].NodeInfo.Address = input.data.NodeInfo.Address
			if input.data.NodeInfo.StartupTime > 0 {
				haHostsDb[hostIndex].NodeInfo.StartupTime = input.data.NodeInfo.StartupTime
			}
			haHostsDb[hostIndex].LastPing = time.Now().Unix()
		}
		return
	}

	if input.remove {
		dbLock.Lock()
		for i, v := range haHostsDb {
			if input.data.NodeInfo.Hostname == v.NodeInfo.Hostname && len(v.NodeInfo.Hostname) > 0 {
				haHostsDb[i] = haHostsDb[len(haHostsDb)-1]
				haHostsDb[len(haHostsDb)-1] = HosterHaNode{}
				haHostsDb = haHostsDb[0 : len(haHostsDb)-1]
				// _ = exec.Command("logger", "-t", "HOSTER_HA_REST", "WARN: host has been removed from the cluster: "+input.data.NodeInfo.Hostname).Run()
				internalLog.Warnf("removed the node from the cluster: %s", input.data.NodeInfo.Hostname)
			}
		}
		return
	}
}

func readHostsDb(dbLock *sync.RWMutex) (db []HosterHaNode) {
	defer func() {
		if r := recover(); r != nil {
			errorValue := fmt.Sprintf("%s", r)
			// _ = exec.Command("logger", "-t", "HOSTER_HA_REST", "PANIC AVOIDED: readHostsDb(): "+errorValue).Run()
			internalLog.Warnf("PANIC AVOIDED: readHostsDb() %s", errorValue)
		}
		// dbLock.RUnlock()
	}()

	dbLock.RLock()
	defer dbLock.RUnlock()
	db = append(db, haHostsDb...)

	return
}

func TerminateOtherMembers() {
	defer func() {
		if r := recover(); r != nil {
			errorValue := fmt.Sprintf("%s", r)
			// _ = exec.Command("logger", "-t", "HOSTER_HA_REST", "PANIC AVOIDED: terminateOtherMembers(): "+errorValue).Run()
			internalLog.Warnf("PANIC AVOIDED: TerminateOtherMembers() %s", errorValue)
		}
	}()

	candidateFound := false
	for _, v := range haConf.Candidates {
		if v.Hostname == myHostname {
			candidateFound = true
		}
	}
	if !candidateFound {
		// _ = exec.Command("logger", "-t", "HOSTER_HA_REST", "ERROR: not a candidate node, use one of the candidates to gracefully stop the RestAPI across the whole cluster").Run()
		internalLog.Errorf("not a candidate node! use one of the candidates to gracefully stop the RestAPI across the whole cluster")
		return
	}

	var wg sync.WaitGroup
	for _, v := range readHostsDb(&hostsDbLock) {
		if v.NodeInfo.Hostname == myHostname {
			continue
		}
		wg.Add(1)
		go func(node HosterHaNode) {
			defer func() {
				if r := recover(); r != nil {
					errorValue := fmt.Sprintf("%s", r)
					// _ = exec.Command("logger", "-t", "HOSTER_HA_REST", "PANIC AVOIDED: terminateOtherMembers()->GR: "+errorValue).Run()
					internalLog.Warnf("PANIC AVOIDED: 544 %s", errorValue)
				}
				wg.Done()
			}()

			// _ = exec.Command("logger", "-t", "HOSTER_HA_REST", "INFO: sending a shutdown signal to: "+node.NodeInfo.Hostname).Run()
			internalLog.Infof("sending a shutdown signal to: %s", node.NodeInfo.Hostname)
			url := node.NodeInfo.Protocol + "://" + node.NodeInfo.Address + ":" + node.NodeInfo.Port + "/api/v2/ha/terminate"

			req, err := http.NewRequest("POST", url, nil)
			if err != nil {
				// _ = exec.Command("logger", "-t", "HOSTER_HA_REST", "ERROR: could not form the /terminate request: "+err.Error()).Run()
				internalLog.Errorf("could not form the /terminate request: %s", err.Error())
				return
			}

			auth := node.NodeInfo.User + ":" + node.NodeInfo.Password
			authEncoded := base64.StdEncoding.EncodeToString([]byte(auth))
			req.Header.Add("Authorization", "Basic "+authEncoded)
			_, err = http.DefaultClient.Do(req)

			if err != nil {
				// _ = exec.Command("logger", "-t", "HOSTER_HA_REST", "WARN: could not notify the member: "+node.NodeInfo.Hostname+". Error: "+err.Error()).Run()
				internalLog.Errorf("could not notify the node about termination ::%s::, exact error ::%s::", node.NodeInfo.Hostname, err.Error())
			}
		}(v)
	}
	wg.Wait()
}
