package main

import (
	"HosterCore/cmd"
	RestApiConfig "HosterCore/internal/app/rest_api_v2/pkg/config"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"time"
)

const timesFailedMax = 3

var timesFailed = 0
var restConf RestApiConfig.RestApiConfig

func init() {
	var err error
	restConf, err = RestApiConfig.GetApiConfig()
	if err != nil {
		logInternal.Panicf("could not read the API config: %s", err.Error())
	}

	if !restConf.HaMode {
		return
	}

	// if restConf.HaDebug {
	// 	_ = exec.Command("logger", "-t", "HOSTER_HA_REST", "DEBUG: hoster_rest_api started in DEBUG mode").Run()
	// } else {
	// 	_ = exec.Command("logger", "-t", "HOSTER_HA_REST", "PROD: hoster_rest_api started in PRODUCTION mode").Run()
	// }

	_ = exec.Command("logger", "-t", "", "DEBUG: hoster_rest_api service start-up").Run()

	// Execute ha_watchdog
	if restConf.HaDebug {
		os.Setenv("REST_API_HA_DEBUG", "true")
	}
	haWatchdogCmd := exec.Command("/opt/hoster-core/ha_watchdog")
	haWatchdogCmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	_ = haWatchdogCmd.Start()

	go func() {
		pingWatchdog()
	}()
}

func pingWatchdog() {
	for {
		time.Sleep(time.Second * 4)

		out, err := exec.Command("pgrep", "ha_watchdog").CombinedOutput()
		if err != nil {
			_ = exec.Command("logger", "-t", "HOSTER_HA_REST", "ERROR: ha_watchdog process is not running").Run()
			timesFailed += 1
		} else {
			_ = exec.Command("kill", "-SIGHUP", strings.TrimSpace(string(out))).Run()
			timesFailed = 0
		}

		if timesFailed >= timesFailedMax {
			if restConf.HaDebug {
				_ = exec.Command("logger", "-t", "HOSTER_HA_REST", "DEBUG: process will exit due to HA_WATCHDOG failure").Run()
				_ = exec.Command("logger", "-t", "HOSTER_HA_REST", "DEBUG: the host system shall reboot soon").Run()
				os.Exit(1)
			} else {
				_ = exec.Command("logger", "-t", "HOSTER_HA_REST", "PROD: process will exit due to HA_WATCHDOG failure").Run()
				cmd.LockAllVms()
				_ = exec.Command("logger", "-t", "HOSTER_HA_REST", "PROD: the host system shall reboot soon").Run()
				_ = exec.Command("reboot").Run()
			}
		}
	}
}
