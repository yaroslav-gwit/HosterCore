package HandlersHA

import (
	RestApiConfig "HosterCore/internal/app/rest_api_v2/pkg/config"
	FreeBSDsysctls "HosterCore/internal/pkg/freebsd/sysctls"
	"os"
	"sync"
	"time"
)

var haConf RestApiConfig.HaConfig
var restConf RestApiConfig.RestApiConfig
var haHostsDb []HosterHaNode
var hostsDbLock sync.RWMutex

func init() {
	var err error
	restConf, err = RestApiConfig.GetApiConfig()
	if err != nil {
		intLog.Panicf("could not read the API config: %s", err.Error())
	}

	if !restConf.HaMode {
		// _ = exec.Command("logger", "-t", "HOSTER_REST", "INFO: STARING REST API SERVER IN REGULAR (NON-HA) MODE").Run()
		intLog.Info("RestAPI server started in a normal mode. HA functionality is disabled.")
		return
	} else {
		intLog.Info("RestAPI server started in the high-availability mode. HA functionality is enabled.")
	}

	if restConf.HaDebug {
		haModeString := []byte("DEBUG")
		_ = os.Remove("/var/run/hoster_rest_api.mode")
		err := os.WriteFile("/var/run/hoster_rest_api.mode", haModeString, 0644)
		if err != nil {
			// _ = exec.Command("logger", "-t", "HOSTER_REST", "ERROR: could not write the /var/run/hoster_rest_api.mode file").Run()
			intLog.Error("could not create the /var/run/hoster_rest_api.mode file")
		}
	} else {
		haModeString := []byte("PRODUCTION")
		_ = os.Remove("/var/run/hoster_rest_api.mode")
		err := os.WriteFile("/var/run/hoster_rest_api.mode", haModeString, 0644)
		if err != nil {
			// _ = exec.Command("logger", "-t", "HOSTER_REST", "ERROR: could not write the /var/run/hoster_rest_api.mode file").Run()
			intLog.Error("could not create the /var/run/hoster_rest_api.mode file")
		}
	}

	haConf, err = RestApiConfig.GetHaConfig()
	if err != nil {
		intLog.Panic(err)
	}

	if haConf.FailOverTime < 10 {
		haConf.FailOverTime = 60
	}
	// _ = exec.Command("logger", "-t", "HOSTER_HA_REST", "INFO: node failover time is: "+strconv.Itoa(int(haConf.FailOverTime))+" seconds").Run()
	intLog.Infof("node failover time is %d seconds", haConf.FailOverTime)

	haConf.StartupTime = time.Now().UnixMilli()

	// Candidates and workers
	go registerNode()
	go trackCandidatesOnline()
	go sendPing()

	// Candidates only
	for _, v := range haConf.Candidates {
		hostname, _ := FreeBSDsysctls.SysctlKernHostname()
		if v.Hostname == hostname {
			go trackManager()
			go removeOfflineNodes()
		}
	}
}
