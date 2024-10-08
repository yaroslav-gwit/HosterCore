package main

import (
	CarpUtils "HosterCore/internal/app/ha_carp/utils"
	"sync"
	"time"
)

var hosts []CarpUtils.HostInfo
var backups []CarpUtils.BackupInfo
var offlineHosts []CarpUtils.HostInfo
var offlineBackups []CarpUtils.BackupInfo

var iAmMaster bool
var becameMaster int64
var masterCheckInterval = 5 * time.Second
var pingRemoteMasterInterval = 10 * time.Second

var mutexHosts = &sync.RWMutex{}
var mutexBackups = &sync.RWMutex{}

func addNewHost(host CarpUtils.HostInfo) {
	mutexHosts.Lock()
	hosts = append(hosts, host)
	mutexHosts.Unlock()
}

func addNewBackup(backup CarpUtils.BackupInfo) {
	mutexBackups.Lock()
	backups = append(backups, backup)
	mutexBackups.Unlock()
}

func listHosts() []CarpUtils.HostInfo {
	mutexHosts.RLock()
	defer mutexHosts.RUnlock()
	return hosts
}

func listBackups() []CarpUtils.BackupInfo {
	mutexBackups.RLock()
	defer mutexBackups.RUnlock()
	return backups
}

func receivePing(host CarpUtils.HostInfo) {
	found := false
	mutexHosts.Lock()
	for i, v := range hosts {
		if v.HostName == host.HostName {
			hosts[i] = host
			hosts[i].LastSeen = time.Now().Local().Unix()
			found = true
		}
	}

	defer mutexHosts.Unlock()
	if !found {
		hosts = append(hosts, host)
	}
}

func removeHost(host CarpUtils.HostInfo) {
	mutexHosts.Lock()
	defer mutexHosts.Unlock()
	for i, h := range hosts {
		if h.HostName == host.HostName {
			hosts = append(hosts[:i], hosts[i+1:]...)
			return
		}
	}
}
