package cmd

import (
	"errors"
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
		Use:   "api-server",
		Short: "Start an API server",
		Long:  `Start an API server on port 3000 (default).`,
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

	apiStartCmd = &cobra.Command{
		Use:   "start",
		Short: "Start an API server",
		Long:  `Start an API server on port 3000 (default).`,
		Run: func(cmd *cobra.Command, args []string) {
			err := checkInitFile()
			if err != nil {
				log.Fatal(err.Error())
			}
			err = startApiServer(apiStartPort, apiStartUser, apiStartPassword, apiHaMode)
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
			out, err := exec.Command("pgrep", "hoster_rest_api").CombinedOutput()
			if err != nil {
				log.Fatal(string(out), err)
			} else {
				fmt.Println("The process is running as PID: " + strings.TrimSpace(string(out)))
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

func startApiServer(port int, user string, password string, haMode bool) error {
	os.Setenv("REST_API_PORT", strconv.Itoa(port))
	os.Setenv("REST_API_USER", user)
	os.Setenv("REST_API_PASSWORD", password)
	if haMode {
		os.Setenv("REST_API_HA_MODE", "true")
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
	emojlog.PrintLogMessage("Using unencrypted/plain HTTP protocol", emojlog.Warning)
	emojlog.PrintLogMessage("HTTP Basic auth username: "+user, emojlog.Info)
	emojlog.PrintLogMessage("HTTP Basic Auth password: "+password, emojlog.Info)

	return nil
}

func stopApiServer() error {
	stdOut, stdErr := exec.Command("pgrep", "-lf", "hoster_rest_api").CombinedOutput()
	if stdErr != nil {
		return errors.New("REST API server is not running")
	}

	reMatch := regexp.MustCompile(`.*hoster_rest_api &.*`)
	reSplit := regexp.MustCompile(`\s+`)

	processId := ""
	for _, v := range strings.Split(string(stdOut), "\n") {
		if reMatch.MatchString(v) {
			processId = reSplit.Split(v, -1)[0]
			break
		}
	}

	_ = exec.Command("kill", "-SIGTERM", processId).Run()
	emojlog.PrintLogMessage("The process has been killed: "+processId, emojlog.Changed)

	out, _ := exec.Command("pgrep", "ha_watchdog").CombinedOutput()
	_ = exec.Command("kill", "-SIGKILL", strings.TrimSpace(string(out)))

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
