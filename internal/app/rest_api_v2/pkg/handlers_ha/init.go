//go:build freebsd
// +build freebsd

package HandlersHA

import (
	RestApiConfig "HosterCore/internal/app/rest_api_v2/pkg/config"
	FreeBSDLogger "HosterCore/internal/pkg/freebsd/logger"
	FreeBSDsysctls "HosterCore/internal/pkg/freebsd/sysctls"
	"os"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// Log Init
var internalLog *logrus.Logger

func init() {
	// Ignore logging if version was requested
	args := os.Args
	if len(args) > 1 {
		res := os.Args[1]
		if res == "version" || res == "v" || res == "--version" || res == "-v" {
			return
		}
	}

	internalLog = logrus.New() // internal HA log to hoster_ha.log, the variable name was set to this in order to avoid any conflicts with the global RestAPI logger
	logFile := HA_LOG_LOCATION

	// Log as JSON instead of the default ASCII/text formatter.
	// log.SetFormatter(&logrus.JSONFormatter{})

	// Output to stdout instead of the default stderr
	internalLog.SetOutput(os.Stdout)

	// Log to file, but fallback to STDOUT if something goes wrong
	file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		FreeBSDLogger.LoggerToSyslog(FreeBSDLogger.LOGGER_SRV_SCHEDULER, FreeBSDLogger.LOGGER_LEVEL_ERROR, "DNS_SERVER: could not use this file for logging "+logFile+", falling back to STDOUT")
	} else {
		internalLog.SetOutput(file)
	}

	internalLog.SetLevel(logrus.DebugLevel)
	// log.SetReportCaller(true)
}

// HA Init
var haConf RestApiConfig.HaConfig
var restConf RestApiConfig.RestApiConfig
var haHostsDb []HosterHaNode
var hostsDbLock sync.RWMutex

func init() {
	// Ignore this init if version was requested
	args := os.Args
	if len(args) > 1 {
		res := os.Args[1]
		if res == "version" || res == "v" || res == "--version" || res == "-v" {
			return
		}
	}

	var err error
	restConf, err = RestApiConfig.GetApiConfig()
	if err != nil {
		internalLog.Panicf("could not read the API config: %s", err.Error())
	}

	if !restConf.HaMode {
		// _ = exec.Command("logger", "-t", "HOSTER_REST", "INFO: STARING REST API SERVER IN REGULAR (NON-HA) MODE").Run()
		internalLog.Info("RestAPI server started in a normal mode. HA functionality is disabled.")
		return
	} else {
		internalLog.Info("RestAPI server started in the high-availability mode. HA functionality is enabled.")
	}

	if restConf.HaDebug {
		haModeString := []byte("DEBUG")
		_ = os.Remove("/var/run/hoster_rest_api.mode")
		err := os.WriteFile("/var/run/hoster_rest_api.mode", haModeString, 0644)
		if err != nil {
			// _ = exec.Command("logger", "-t", "HOSTER_REST", "ERROR: could not write the /var/run/hoster_rest_api.mode file").Run()
			internalLog.Error("could not create the /var/run/hoster_rest_api.mode file")
		}
	} else {
		haModeString := []byte("PRODUCTION")
		_ = os.Remove("/var/run/hoster_rest_api.mode")
		err := os.WriteFile("/var/run/hoster_rest_api.mode", haModeString, 0644)
		if err != nil {
			// _ = exec.Command("logger", "-t", "HOSTER_REST", "ERROR: could not write the /var/run/hoster_rest_api.mode file").Run()
			internalLog.Error("could not create the /var/run/hoster_rest_api.mode file")
		}
	}

	haConf, err = RestApiConfig.GetHaConfig()
	if err != nil {
		internalLog.Panic(err)
	}

	if haConf.FailOverTime < 10 {
		haConf.FailOverTime = 60
	}
	// _ = exec.Command("logger", "-t", "HOSTER_HA_REST", "INFO: node failover time is: "+strconv.Itoa(int(haConf.FailOverTime))+" seconds").Run()
	internalLog.Infof("node failover time is %d seconds", haConf.FailOverTime)

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
