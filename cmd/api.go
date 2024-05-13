//go:build freebsd
// +build freebsd

package cmd

import (
	RestApiConfig "HosterCore/internal/app/rest_api_v2/pkg/config"
	"HosterCore/internal/pkg/emojlog"
	FreeBSDKill "HosterCore/internal/pkg/freebsd/kill"
	FreeBSDPgrep "HosterCore/internal/pkg/freebsd/pgrep"
	HosterLocations "HosterCore/internal/pkg/hoster/locations"
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
		Short: "RestAPI Service",
		Long:  `RestAPI Service controls.`,
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
			// err := startApiServer()

			pid, err := FreeBSDPgrep.FindRestAPIv2()
			if err == nil || pid != 0 {
				emojlog.PrintLogMessage("RestAPIv2 server is already running", emojlog.Error)
				os.Exit(1)
			}

			bin, err := HosterLocations.LocateBinary("rest_api_v2")
			if err != nil {
				emojlog.PrintLogMessage(err.Error(), emojlog.Error)
				os.Exit(1)
			}
			config, err := HosterLocations.LocateConfig("restapi_config.json")
			if err != nil {
				emojlog.PrintLogMessage(err.Error(), emojlog.Error)
				os.Exit(1)
			}
			file, err := os.ReadFile(config)
			if err != nil {
				emojlog.PrintLogMessage(err.Error(), emojlog.Error)
				os.Exit(1)
			}

			restApiConfig := RestApiConfig.RestApiConfig{}
			err = json.Unmarshal(file, &restApiConfig)
			if err != nil {
				emojlog.PrintLogMessage(err.Error(), emojlog.Error)
				os.Exit(1)
			}

			os.Setenv("LOG_FILE", "/var/log/hoster_rest_api_v2.log")
			command := exec.Command(bin)
			command.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
			err = command.Start()
			if err != nil {
				emojlog.PrintLogMessage(err.Error(), emojlog.Error)
				os.Exit(1)
			}

			emojlog.PrintLogMessage("Started the REST API server on port: "+strconv.Itoa(restApiConfig.Port), emojlog.Info)
			emojlog.PrintLogMessage("You can find user credentials inside of this config file: "+config, emojlog.Info)
			if restApiConfig.Protocol != "https" {
				emojlog.PrintLogMessage("Using unencrypted/plain HTTP protocol (don't forget to encapsulate it within the WireGuard tunnel)", emojlog.Warning)
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
			// err := StopApiServer()

			apiPid, _ := FreeBSDPgrep.FindRestAPIv2()
			watchdogPid, _ := FreeBSDPgrep.FindWatchdog()

			if watchdogPid > 0 {
				err1 := FreeBSDKill.KillProcess(FreeBSDKill.KillSignalINT, apiPid)
				err2 := FreeBSDKill.KillProcess(FreeBSDKill.KillSignalTERM, watchdogPid)

				if err1 != nil {
					emojlog.PrintLogMessage(err1.Error(), emojlog.Error)
					os.Exit(1)
				}
				if err2 != nil {
					emojlog.PrintLogMessage(err2.Error(), emojlog.Error)
					os.Exit(1)
				}
				os.Exit(0)
			} else if apiPid > 0 {
				err := FreeBSDKill.KillProcess(FreeBSDKill.KillSignalTERM, apiPid)
				if err != nil {
					emojlog.PrintLogMessage(err.Error(), emojlog.Error)
					os.Exit(1)
				}
				os.Exit(0)
			}

			emojlog.PrintLogMessage("RestAPIv2 service is not running", emojlog.Error)
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
			// err := statusApiServer()

			debug, prod := getHaRunMode()
			apiPid, _ := FreeBSDPgrep.FindRestAPIv2()
			wdPid, _ := FreeBSDPgrep.FindWatchdog()

			if apiPid > 0 {
				fmt.Println(" 🟢 API Server is running as PID " + strconv.Itoa(apiPid))
			} else {
				fmt.Println(" 🔴 API Server IS NOT running")
			}

			if wdPid > 0 {
				fmt.Println(" 🟢 HA Watchdog service is running as PID " + strconv.Itoa(wdPid))
				fmt.Println()

				if debug {
					fmt.Println(" ️🤖 HA is running in DEBUG mode")
				}
				if prod {
					fmt.Println(" 🚀 HA is running in PRODUCTION mode")
				}
				fmt.Println(" 🔶 BE CAREFUL! This system is running as a part of the HA. Changes applied here, may affect other cluster members.")
			} else {
				fmt.Println(" 🔴 HA Watchdog service IS NOT running")
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

type RestApiConfigLocal struct {
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

	restApiConfig := RestApiConfigLocal{}
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
			err := FreeBSDKill.KillProcess(FreeBSDKill.KillSignalINT, services.ApiServerPid)
			if err != nil {
				return err
			}
		} else {
			err := FreeBSDKill.KillProcess(FreeBSDKill.KillSignalTERM, services.ApiServerPid)
			if err != nil {
				return err
			}
		}
		emojlog.PrintLogMessage("ha_watchdog service has been stopped using PID: "+strconv.Itoa(services.ApiServerPid), emojlog.Changed)
	}

	if services.HaWatchdogRunning {
		err := FreeBSDKill.KillProcess(FreeBSDKill.KillSignalTERM, services.HaWatchDogPid)
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
	// tailCmd := exec.Command("tail", "-35", "-f", "/var/run/hoster_rest_api.log")
	// tailCmd := exec.Command("tail", "-35", "-f", "/var/log/hoster_rest_api.log")
	tailCmd := exec.Command("tail", "-35", "-f", "/var/log/hoster_rest_api_v2.log")

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
	ApiServerRunning  bool
	HaWatchDogPid     int
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
