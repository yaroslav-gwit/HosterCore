package main

import (
	"HosterCore/cmd"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"
)

var version = "" // version is set by the build system

func main() {
	// Print the version and exit
	args := os.Args
	if len(args) > 1 {
		res := os.Args[1]
		if res == "version" || res == "v" || res == "--version" || res == "-v" {
			fmt.Println(version)
			return
		}
	}

	log.Info("new service start-up")

	var debugMode bool
	debugModeEnv := os.Getenv("REST_API_HA_DEBUG")
	if len(debugModeEnv) > 0 {
		debugMode = true
	}

	if debugMode {
		log.Info("started in DEBUG mode")
	} else {
		log.Info("started in PROD mode")
	}

	timesFailed := 0
	timesFailedMax := 2
	lastReachOut := time.Now().Unix()

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGHUP, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		for sig := range signals {
			if sig == syscall.SIGHUP {
				lastReachOut = time.Now().Unix()
			}
			if sig == syscall.SIGTERM || sig == syscall.SIGINT {
				log.Info("received SIGTERM, executing a graceful exit")
				os.Exit(0)
			}
		}
	}()

	for {
		time.Sleep(time.Second * 5)

		if timesFailed >= timesFailedMax {
			if debugMode {
				log.Debug("(DEBUG) RestAPI process has failed, rebooting the system now")
				os.Exit(1)
			} else {
				log.Warn("RestAPI process has failed, rebooting the system now")
				cmd.LockAllVms()
				_ = exec.Command("reboot").Run()
			}
		}

		if time.Now().Unix() > lastReachOut+5 {
			log.Debugf("ping failed, previous alive timestamp: %d", lastReachOut)
			timesFailed += 1
			log.Debugf("pings missed so far: %d, may need to fence the system after this many failed pings: %d", timesFailed, timesFailedMax)
		} else {
			timesFailed = 0
		}
	}
}
