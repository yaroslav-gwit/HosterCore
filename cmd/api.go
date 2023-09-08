package cmd

import (
	"fmt"
	"hoster/emojlog"
	"log"
	"os"
	"os/exec"
	"path"
	"regexp"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

var (
	apiCmd = &cobra.Command{
		Use:   "api",
		Short: "API Server Menu",
		Long:  `API Server Menu.`,
		Run: func(cmd *cobra.Command, args []string) {
			err := checkInitFile()
			if err != nil {
				log.Fatal(err.Error())
			}
			cmd.Help()
		},
	}
)

var (
	apiStartPort     int
	apiStartUser     string
	apiStartPassword string
	apiHaMode        bool
	apiHaDebug       bool

	apiStartCmd = &cobra.Command{
		Use:   "start",
		Short: "Start an API server",
		Long:  `Start an API server on port 3000 (default).`,
		Run: func(cmd *cobra.Command, args []string) {
			err := checkInitFile()
			if err != nil {
				log.Fatal(err.Error())
			}
			err = startApiServer(apiStartPort, apiStartUser, apiStartPassword, apiHaMode, apiHaDebug)
			if err != nil {
				log.Fatal(err.Error())
			}
		},
	}
)

var (
	apiStopCmd = &cobra.Command{
		Use:   "stop",
		Short: "Stop the API server",
		Long:  `Stop the API server.`,
		Run: func(cmd *cobra.Command, args []string) {
			err := checkInitFile()
			if err != nil {
				log.Fatal(err.Error())
			}
			err = stopApiServer()
			if err != nil {
				log.Fatal(err.Error())
			}
		},
	}
)

var (
	apiStatusCmd = &cobra.Command{
		Use:   "status",
		Short: "Display status of the API server",
		Long:  `Display status of the API server.`,
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			err := checkInitFile()
			if err != nil {
				log.Fatal(err.Error())
			}

			err = statusApiServer()
			if err != nil {
				log.Fatal(err.Error())
			}
		},
	}
)

var (
	apiShowLogCmd = &cobra.Command{
		Use:   "show-log",
		Short: "Show log for the API server",
		Long:  `Show log for the API server.`,
		Run: func(cmd *cobra.Command, args []string) {
			err := checkInitFile()
			if err != nil {
				log.Fatal(err.Error())
			}
			err = showLogApiServer()
			if err != nil {
				log.Fatal(err.Error())
			}
		},
	}
)

func startApiServer(port int, user string, password string, haMode bool, haDebug bool) error {
	os.Setenv("REST_API_PORT", strconv.Itoa(port))
	os.Setenv("REST_API_USER", user)
	os.Setenv("REST_API_PASSWORD", password)
	if haMode {
		os.Setenv("REST_API_HA_MODE", "true")
	}
	if haDebug {
		os.Setenv("REST_API_HA_DEBUG", "true")
	}

	execPath, err := os.Executable()
	if err != nil {
		return err
	}

	execFile := path.Dir(execPath) + "/hoster_rest_api"
	err = exec.Command("nohup", execFile, "&").Start()
	if err != nil {
		return err
	}

	emojlog.PrintLogMessage("Started the REST API server on port: "+strconv.Itoa(port), emojlog.Info)
	emojlog.PrintLogMessage("Using unencrypted/plain HTTP protocol (make sure to put it behind Traefik with SSL termination if the untrusted networks are involved)", emojlog.Warning)
	emojlog.PrintLogMessage("HTTP Basic auth username: "+user, emojlog.Info)
	emojlog.PrintLogMessage("HTTP Basic Auth password: "+password, emojlog.Info)

	return nil
}

func stopApiServer() error {
	timesKilled := 0

	out, err := exec.Command("pgrep", "ha_watchdog").CombinedOutput()
	if err == nil && len(string(out)) > 0 {
		timesKilled += 1
		_ = exec.Command("kill", "-SIGTERM", strings.TrimSpace(string(out))).Run()
		emojlog.PrintLogMessage("HA_WATCHDOG service stopped using PID: "+strings.TrimSpace(string(out)), emojlog.Changed)
	}

	out, err = exec.Command("pgrep", "hoster_rest_api").CombinedOutput()
	if err == nil && len(string(out)) > 0 {
		timesKilled += 1
		_ = exec.Command("kill", "-SIGTERM", strings.TrimSpace(string(out))).Run()
		emojlog.PrintLogMessage("REST API service stopped using PID: "+strings.TrimSpace(string(out)), emojlog.Changed)
	}

	if timesKilled < 1 {
		emojlog.PrintLogMessage("Sorry, the REST API service is not running", emojlog.Error)
	}

	return nil
}

func showLogApiServer() error {
	tailCmd := exec.Command("tail", "-35", "-f", "/var/run/hoster_rest_api.log")

	tailCmd.Stdin = os.Stdin
	tailCmd.Stdout = os.Stdout
	tailCmd.Stderr = os.Stderr

	err := tailCmd.Run()
	if err != nil {
		return err
	}

	return nil
}

func statusApiServer() error {
	apiProcessPgrep := apiServerServiceInfo()

	if apiProcessPgrep.ApiServerRunning {
		fmt.Println(" 🟢 API Server is running as PID " + apiProcessPgrep.ApiServerPid)
	} else {
		fmt.Println(" 🔴 API Server IS NOT running")
	}

	if apiProcessPgrep.HaWatchdogRunning {
		fmt.Println(" 🟢 HA Watchdog service is running as PID " + apiProcessPgrep.HaWatchDogPid)
		fmt.Println(" 🔶 BE CAREFUL! This system is running as a part of the HA. Changes applied here, may affect other cluster members.")
	} else {
		fmt.Println(" 🔴 HA Watchdog service IS NOT running")
	}

	return nil
}

type ApiProcessServiceInfo struct {
	HaWatchdogRunning bool
	HaWatchDogPid     string
	ApiServerRunning  bool
	ApiServerPid      string
}

func apiServerServiceInfo() (pgrepOutput ApiProcessServiceInfo) {
	apiPgrepOut, _ := exec.Command("pgrep", "-lf", "hoster_rest_api").CombinedOutput()
	watchDogPgrepOut, _ := exec.Command("pgrep", "-lf", "ha_watchdog").CombinedOutput()

	reSplitSpace := regexp.MustCompile(`\s+`)
	reMatchApiProcess := regexp.MustCompile(`hoster_rest_api`)
	reMatchHaWatchdogProcess := regexp.MustCompile(`ha_watchdog`)

	reMatchSkipLogProcess := regexp.MustCompile(`.*hoster_rest_api.log`)

	for _, v := range strings.Split(string(apiPgrepOut), "\n") {
		if reMatchApiProcess.MatchString(v) {
			if reMatchSkipLogProcess.MatchString(v) {
				continue
			} else {
				pid := reSplitSpace.Split(v, -1)[0]
				pgrepOutput.ApiServerPid = pid
				pgrepOutput.ApiServerRunning = true
			}
		}
	}

	for _, v := range strings.Split(string(watchDogPgrepOut), "\n") {
		if reMatchHaWatchdogProcess.MatchString(v) {
			pid := reSplitSpace.Split(v, -1)[0]
			pgrepOutput.HaWatchDogPid = pid
			pgrepOutput.HaWatchdogRunning = true
		}
	}

	return
}
