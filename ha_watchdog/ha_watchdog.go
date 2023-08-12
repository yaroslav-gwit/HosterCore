package main

import (
	"hoster/cmd"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"syscall"
	"time"
)

func main() {
	_ = exec.Command("logger", "-t", "HOSTER_HA_WATCHDOG", "service start-up").Run()

	var debugMode bool
	debugModeEnv := os.Getenv("REST_API_HA_DEBUG")
	if len(debugModeEnv) > 0 {
		debugMode = true
	}

	if debugMode {
		_ = exec.Command("logger", "-t", "HOSTER_HA_REST", "DEBUG: HA_WATCHDOG started in DEBUG mode").Run()
	} else {
		_ = exec.Command("logger", "-t", "HOSTER_HA_REST", "HA_WATCHDOG started in PRODUCTION mode").Run()
	}

	timesFailed := 0
	timesFailedMax := 10
	lastReachOut := time.Now().Unix()

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGHUP, syscall.SIGTERM)
	go func() {
		for sig := range signals {
			if sig == syscall.SIGHUP {
				lastReachOut = time.Now().Unix()
			}
			if sig == syscall.SIGTERM {
				_ = exec.Command("logger", "-t", "HOSTER_HA_WATCHDOG", "received SIGTERM, exiting").Run()
				os.Exit(0)
			}
		}
	}()

	for {
		time.Sleep(time.Second * 5)

		if timesFailed >= timesFailedMax {
			if debugMode {
				_ = exec.Command("logger", "-t", "HOSTER_HA_WATCHDOG", "DEBUG: rebooting the system due to failed HA state, pings failed: "+strconv.Itoa(timesFailed)).Run()
				os.Exit(1)
			} else {
				_ = exec.Command("logger", "-t", "HOSTER_HA_WATCHDOG", "PROD: rebooting the system due to failed HA state, pings failed: "+strconv.Itoa(timesFailed)).Run()
				cmd.LockAllVms()
				_ = exec.Command("reboot").Run()
			}
		}

		if time.Now().Unix() > lastReachOut+5 {
			_ = exec.Command("logger", "-t", "HOSTER_HA_WATCHDOG", "ping failed, previous alive timestamp: "+strconv.Itoa(int(lastReachOut))).Run()
			timesFailed += 1
			_ = exec.Command("logger", "-t", "HOSTER_HA_WATCHDOG", "pings missed so far: "+strconv.Itoa(timesFailed)+"; will terminate the system at: "+strconv.Itoa(timesFailedMax)).Run()
		} else {
			timesFailed = 0
		}
	}
}
