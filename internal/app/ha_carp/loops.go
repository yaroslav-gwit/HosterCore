package main

import (
	CarpUtils "HosterCore/internal/app/ha_carp/utils"
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
			iAmMaster = true
			becameMaster = time.Now().Local().Unix()
			return
		}
	}

	iAmMaster = false
}

func pingMaster() {
	// Send a ping to the master node
}

func syncBackupsFromMaster() {
	// Sync the backups with the master node
}

func syncHostsFromMaster() {
	// Sync the backups with the master node
}

func collectBackups() {
	// Collect backups from all nodes
}
