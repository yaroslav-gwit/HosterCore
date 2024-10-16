package main

import (
	CarpUtils "HosterCore/internal/app/ha_carp/utils"
	ApiV2client "HosterCore/internal/pkg/api_v2_client"
	FreeBSDsysctls "HosterCore/internal/pkg/freebsd/sysctls"
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
			hostname, _ := FreeBSDsysctls.SysctlKernHostname()
			iAmMaster = true
			currentMaster = hostname
			becameMaster = time.Now().Local().Unix()
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
		result = append(result, tmp...)
	}

	mutexBackups.Lock()
	backups = result
	mutexBackups.Unlock()
}
