package main

import (
	CarpUtils "HosterCore/internal/app/ha_carp/utils"
	"sync"
)

var hosts []CarpUtils.HostInfo
var backups []CarpUtils.BackupInfo

var mutexHosts = &sync.RWMutex{}
var mutexBackups = &sync.RWMutex{}
