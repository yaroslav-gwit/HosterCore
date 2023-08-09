package main

import (
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"syscall"
	"time"
)

func main() {
	timesFailed := 0
	timesFailedMax := 10
	lastReachOut := time.Now().Unix()

	signals := make(chan os.Signal, 1)
	signal.Notify(signals)
	go func() {
		for sig := range signals {
			if sig == syscall.SIGHUP {
				lastReachOut = time.Now().Unix()
				_ = exec.Command("logger", "-t", "HOSTER_HA_WATCHDOG", "REST API is still alive: "+strconv.Itoa(int(lastReachOut)))
			}
			if sig == syscall.SIGKILL || sig == syscall.SIGTERM {
				_ = exec.Command("logger", "-t", "HOSTER_HA_WATCHDOG", "received kill or term signal, exiting")
				os.Exit(0)
			}
		}
	}()

	for {
		time.Sleep(time.Second * 5)

		if timesFailed >= timesFailedMax {
			_ = exec.Command("logger", "-t", "HOSTER_HA_WATCHDOG", "rebooting the system due to failed HA state, pings failed (number of times): "+strconv.Itoa(timesFailed))
			os.Exit(0)
		}

		if time.Now().Unix() > lastReachOut+5 {
			_ = exec.Command("logger", "-t", "HOSTER_HA_WATCHDOG", "ping failed, previous alive timestamp: "+strconv.Itoa(int(lastReachOut)))
			timesFailed += 1
			_ = exec.Command("logger", "-t", "HOSTER_HA_WATCHDOG", "pings missed so far: "+strconv.Itoa(timesFailed)+"; will terminate the system at: "+strconv.Itoa(timesFailedMax))
		} else {
			timesFailed = 0
		}
	}
}
