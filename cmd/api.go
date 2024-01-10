package cmd

import (
	"HosterCore/emojlog"
	"HosterCore/osfreebsd"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"syscall"

	"github.com/spf13/cobra"
)

var (
	apiCmd = &cobra.Command{
		Use:   "api",
		Short: "API Server Menu",
		Long:  `API Server Menu.`,
		Run: func(cmd *cobra.Command, args []string) {
			checkInitFile()
			cmd.Help()
		},
	}
)

var (
	// apiStartPort     int
	// apiStartUser     string
	// apiStartPassword string
	// apiHaMode        bool
	// apiHaDebug       bool

	apiStartCmd = &cobra.Command{
		Use:   "start",
		Short: "Start an API server",
		Long:  `Start an API server on port 3000 (default).`,
		Run: func(cmd *cobra.Command, args []string) {
			checkInitFile()
			err := startApiServer()
			if err != nil {
				emojlog.PrintLogMessage(err.Error(), emojlog.Error)
				os.Exit(1)
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
			checkInitFile()
			err := StopApiServer()
			if err != nil {
				emojlog.PrintLogMessage(err.Error(), emojlog.Error)
				os.Exit(1)
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
			checkInitFile()
			err := statusApiServer()
			if err != nil {
				emojlog.PrintLogMessage(err.Error(), emojlog.Error)
				os.Exit(1)
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
			checkInitFile()

			err := showLogApiServer()
			if err != nil {
				emojlog.PrintLogMessage(err.Error(), emojlog.Error)
				os.Exit(1)
			}
		},
	}
)

type RestApiConfig struct {
	Bind     string `json:"bind"`
	Port     int    `json:"port"`
	Protocol string `json:"protocol"`
	HaMode   bool   `json:"ha_mode"`
	HaDebug  bool   `json:"ha_debug,omitempty"`
	HTTPAuth []struct {
		User     string `json:"user"`
		Password string `json:"password"`
		HaUser   bool   `json:"ha_user"`
	} `json:"http_auth"`
}

// func startApiServer(port int, user string, password string, haMode bool, haDebug bool) error {
func startApiServer() error {
	// os.Setenv("REST_API_PORT", strconv.Itoa(port))
	// os.Setenv("REST_API_USER", user)
	// os.Setenv("REST_API_PASSWORD", password)
	// if haMode {
	// 	os.Setenv("REST_API_HA_MODE", "true")
	// }
	// if haDebug {
	// 	os.Setenv("REST_API_HA_DEBUG", "true")
	// }

	execPath, err := os.Executable()
	if err != nil {
		return err
	}
	execDir := filepath.Dir(execPath)

	file, err := os.ReadFile(execDir + "/config_files/restapi_config.json")
	if err != nil {
		return errors.New("could not open restapi_config.json: " + err.Error())
	}

	restApiConfig := RestApiConfig{}
	err = json.Unmarshal(file, &restApiConfig)
	if err != nil {
		return errors.New("could not parse restapi_config.json: " + err.Error())
	}

	execFile := path.Dir(execPath) + "/hoster_rest_api"
	cmd := exec.Command(execFile)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	err = cmd.Start()
	if err != nil {
		return err
	}

	emojlog.PrintLogMessage("Started the REST API server on port: "+strconv.Itoa(restApiConfig.Port), emojlog.Info)
	emojlog.PrintLogMessage("Check user credentials at "+execPath+"/config_files/restapi_config.json", emojlog.Info)
	if restApiConfig.Protocol != "https" {
		emojlog.PrintLogMessage("Using unencrypted/plain HTTP protocol (make sure to put it behind Traefik with SSL termination if the untrusted networks are involved)", emojlog.Warning)
	}

	return nil
}

func StopApiServer() error {
	services := ApiProcessServiceInfo()

	if services.ApiServerRunning {
		if services.HaWatchdogRunning {
			err := osfreebsd.KillProcess(osfreebsd.KillSignalINT, services.ApiServerPid)
			if err != nil {
				return err
			}
		} else {
			err := osfreebsd.KillProcess(osfreebsd.KillSignalTERM, services.ApiServerPid)
			if err != nil {
				return err
			}
		}
		emojlog.PrintLogMessage("ha_watchdog service has been stopped using PID: "+strconv.Itoa(services.ApiServerPid), emojlog.Changed)
	}

	if services.HaWatchdogRunning {
		err := osfreebsd.KillProcess(osfreebsd.KillSignalTERM, services.HaWatchDogPid)
		if err != nil {
			return err
		}
		emojlog.PrintLogMessage("hoster_rest_api service has been stopped using PID: "+strconv.Itoa(services.HaWatchDogPid), emojlog.Changed)
	}

	if !services.ApiServerRunning {
		emojlog.PrintLogMessage("Sorry, the REST API service is not running", emojlog.Error)
	}

	return nil
}

func showLogApiServer() error {
	tailCmd := exec.Command("tail", "-35", "-f", "/var/log/hoster_rest_api.log")
	// tailCmd := exec.Command("tail", "-35", "-f", "/var/run/hoster_rest_api.log")

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
	apiProcessPgrep := ApiProcessServiceInfo()
	debugMode, productionMode := getHaRunMode()

	if apiProcessPgrep.ApiServerRunning {
		fmt.Println(" 🟢 API Server is running as PID " + strconv.Itoa(apiProcessPgrep.ApiServerPid))
	} else {
		fmt.Println(" 🔴 API Server IS NOT running")
	}

	if apiProcessPgrep.HaWatchdogRunning {
		fmt.Println(" 🟢 HA Watchdog service is running as PID " + strconv.Itoa(apiProcessPgrep.HaWatchDogPid))
		fmt.Println()

		if debugMode {
			fmt.Println(" ️🤖 HA is running in DEBUG mode")
		}
		if productionMode {
			fmt.Println(" 🚀 HA is running in PRODUCTION mode")
		}

		fmt.Println(" 🔶 BE CAREFUL! This system is running as a part of the HA. Changes applied here, may affect other cluster members.")
	} else {
		fmt.Println(" 🔴 HA Watchdog service IS NOT running")
	}

	return nil
}

type ApiProcessServiceInfoStruct struct {
	HaWatchdogRunning bool
	HaWatchDogPid     int
	ApiServerRunning  bool
	ApiServerPid      int
}

func ApiProcessServiceInfo() (pgrepOutput ApiProcessServiceInfoStruct) {
	apiPgrepOut, _ := exec.Command("pgrep", "-lf", "hoster_rest_api").CombinedOutput()
	watchDogPgrepOut, _ := exec.Command("pgrep", "-lf", "ha_watchdog").CombinedOutput()

	reSplitSpace := regexp.MustCompile(`\s+`)
	reMatchApiProcess := regexp.MustCompile(`.*hoster_rest_api.*`)
	reMatchHaWatchdogProcess := regexp.MustCompile(`.*ha_watchdog.*`)

	reMatchSkipLogProcess := regexp.MustCompile(`.*hoster_rest_api\.log.*`)

	for _, v := range strings.Split(string(apiPgrepOut), "\n") {
		if reMatchApiProcess.MatchString(v) {
			if reMatchSkipLogProcess.MatchString(v) {
				continue
			} else {
				pidStr := reSplitSpace.Split(v, -1)[0]
				pid, err := strconv.Atoi(strings.TrimSpace(pidStr))
				if err != nil {
					pid = 0
					pgrepOutput.ApiServerRunning = false
				} else {
					pgrepOutput.ApiServerRunning = true
				}
				pgrepOutput.ApiServerPid = pid
			}
		}
	}

	for _, v := range strings.Split(string(watchDogPgrepOut), "\n") {
		if reMatchHaWatchdogProcess.MatchString(v) {
			pidStr := reSplitSpace.Split(v, -1)[0]
			pid, err := strconv.Atoi(strings.TrimSpace(pidStr))
			if err != nil {
				pid = 0
				pgrepOutput.HaWatchdogRunning = false
			} else {
				pgrepOutput.HaWatchdogRunning = true
			}
			pgrepOutput.HaWatchDogPid = pid
		}
	}

	return
}

func getHaRunMode() (haRunInDebug bool, haRunInProduction bool) {
	fileData, err := os.ReadFile("/var/run/hoster_rest_api.mode")
	if err != nil {
		return
	}

	reMatchProduction := regexp.MustCompile(`.*[Pp][Rr][Oo][Dd].*`)
	if reMatchProduction.MatchString(string(fileData)) {
		haRunInProduction = true
		return
	}

	reMatchDebug := regexp.MustCompile(`.*[Dd][Ee][Bb][Uu][Gg].*`)
	if reMatchDebug.MatchString(string(fileData)) {
		haRunInDebug = true
		return
	}

	return
}
