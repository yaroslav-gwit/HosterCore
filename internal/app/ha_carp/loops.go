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
		log.Debug("Not master, skipping sync")
		return
	}

	ha := CarpUtils.HaStatus{}
	ha.Resources = listBackups()
	ha.Hosts = listHosts()

	hostname, _ := FreeBSDsysctls.SysctlKernHostname()
	wg := sync.WaitGroup{}
	for _, v := range ha.Hosts {
		wg.Add(1)
		log.Debug("Sending local state to", v.IpAddress)
		go func(v CarpUtils.HostInfo, wg *sync.WaitGroup) {
			defer wg.Done()
			if hostname == currentMaster { // Don't send the state to self
				return
			}

			err := ApiV2client.SendLocalState(ha, v.IpAddress, currentMaster)
			if err != nil {
				log.Errorf("Error sending local state to %s: %s", v.IpAddress, err.Error())
			}
			log.Debug("Sent local state to", v.IpAddress)
		}(v, &wg)
	}

	wg.Wait()
}

func syncHosts() {
	// Sync the backups with the master node
}

func collectBackups() {
	// Collect backups from all nodes
}
