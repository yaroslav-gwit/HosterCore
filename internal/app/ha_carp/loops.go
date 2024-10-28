package main

import (
	CarpUtils "HosterCore/internal/app/ha_carp/utils"
	ApiV2client "HosterCore/internal/pkg/api_v2_client"
	FreeBSDsysctls "HosterCore/internal/pkg/freebsd/sysctls"
	"slices"
	"sort"
	"strings"
	"sync"
	"time"
)

func checkIfMaster() {
	iface, err := CarpUtils.ParseIfconfig()
	if err != nil {
		log.Error("Error parsing ifconfig:", err)
		return
	}

	// Check if the interface is in MASTER state
	for _, v := range iface {
		if v.Status == "MASTER" {
			if !iAmMaster {
				// If we just became master, update the timestamp
				becameMaster = time.Now().Local().Unix()

				// Clear the offline status for all hosts when we become master
				// This is required to re-trigger the failover process for the hosts that are still offline
				mutexHosts.Lock()
				for i := range hosts {
					hosts[i].Offline = false
				}
				mutexHosts.Unlock()
			}

			hostname, _ := FreeBSDsysctls.SysctlKernHostname()
			iAmMaster = true
			currentMaster = hostname

			return
		}
	}

	iAmMaster = false
}

func pingMaster() {
	hostname, err := ApiV2client.PingMaster(activeHaConfig)
	if err != nil {
		log.Error("Error pinging master:", err)
		return
	}

	if hostname != currentMaster {
		currentMaster = hostname
	}
}

func syncState() {
	if !iAmMaster {
		log.Debug("STATE SYNC: I am not a master, skipping fan-out sync")
		return
	}

	ha := CarpUtils.HaStatus{}
	ha.Resources = listBackups()
	ha.Hosts = listHosts()

	wg := sync.WaitGroup{}
	log.Debug("STATE SYNC: Begin syncing state using fan-out")
	for _, v := range ha.Hosts {
		wg.Add(1)
		log.Debug("STATE SYNC: Sending local state to ", v.IpAddress)

		go func(v CarpUtils.HostInfo, wg *sync.WaitGroup) {
			defer wg.Done()

			err := ApiV2client.SendLocalState(ha, v.IpAddress, currentMaster)
			if err != nil {
				log.Errorf("STATE SYNC: Error sending local state to %s: %s", v.IpAddress, err.Error())
			}

			log.Debug("STATE SYNC: Sent local state to ", v.IpAddress)
		}(v, &wg)
	}

	wg.Wait()
	log.Debug("STATE SYNC: Done syncing state using fan-out")
}

func getRemoteBackups() {
	result := []CarpUtils.BackupInfo{}

	if !iAmMaster {
		log.Debug("BACKUP SYNC: I am not a master, skipping backup sync")
		return
	}

	for _, v := range listHosts() { // Get backups from all hosts (naive approach, for now)
		tmp, err := ApiV2client.ReturnBackups(v)
		if err != nil {
			log.Errorf("Error getting backups from %s: %s", v.IpAddress, err.Error())
			continue
		}

		for i := range tmp {
			tmp[i].BasePayload = CarpUtils.BasePayload{}
			tmp[i].Type = ""
		}

		result = append(result, tmp...)
	}

	mutexBackups.Lock()
	backups = []CarpUtils.BackupInfo{}
	backups = append(backups, result...)
	mutexBackups.Unlock()
}

func detectOfflineHosts() {
	if isSelfMaster() {
		return
	}

	mutexHosts.Lock()
	config, err := CarpUtils.ParseCarpConfigFile()
	if err != nil {
		log.Error("Error getting config: ", err)
		return
	}

	for i, v := range hosts {
		if v.Offline {
			continue
		}

		if time.Now().Local().Unix()-v.LastSeen > int64(config.FailoverAfter) { // Remove hosts that haven't been seen in a while
			hosts[i].Offline = true
			addOfflineBackup(v.HostName)
			log.Warnf("Host %s has gone offline", v.HostName)
		}
	}

	mutexHosts.Unlock()
}

func failOverResource() {
	if failoverInProcess {
		return
	}

	if !isSelfMaster() {
		return
	}
	if len(offlineBackups) < 1 {
		return
	}

	failoverInProcess = true
	mutexOfflineBackups.Lock()
	defer func() {
		failoverInProcess = false
		mutexOfflineBackups.Unlock()
	}()

	if len(offlineBackups) == 1 {
		log.Warnf("Failing over resource %s to %s", offlineBackups[0].ResourceName, offlineBackups[0].ParentHost)
		return
	}

	// Sort by ResourceName
	sort.SliceStable(offlineBackups, func(i, j int) bool {
		return strings.ToLower(backups[i].ResourceName) < strings.ToLower(backups[j].ResourceName)
	})

	localList := []CarpUtils.BackupInfo{}
	for _, v := range offlineBackups { // filter out the latest snapshot for each resource
		for ii, vv := range localList {
			if slices.Contains(localList, v) {
				continue
			}

			if v.ResourceName == vv.ResourceName {
				if v.LastSnapshot > vv.LastSnapshot {
					localList[ii] = v
				}
				continue
			}

			localList = append(localList, v)
		}
	}

	failed := []CarpUtils.BackupInfo{}
	for _, v := range localList {
		log.Warnf("Failing over resource %s to %s", v.ResourceName, v.ParentHost)
	}

	offlineBackups = []CarpUtils.BackupInfo{}
	offlineBackups = append(offlineBackups, failed...)
}
