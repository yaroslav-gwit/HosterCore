package cmd

import (
	"errors"
	"hoster/emojlog"
	"log"
	"os"
	"os/exec"
	"path"
	"regexp"
	"strings"

	"github.com/spf13/cobra"
)

var (
	dnsCmd = &cobra.Command{
		Use:   "dns",
		Short: "Hoster integrated DNS Server",
		Long:  `Hoster integrated DNS Server`,
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
	dnsStartCmd = &cobra.Command{
		Use:   "start",
		Short: "Initialize Hoster integrated DNS Server",
		Long:  `Initialize Hoster integrated DNS Server`,
		Run: func(cmd *cobra.Command, args []string) {
			err := checkInitFile()
			if err != nil {
				log.Fatal(err.Error())
			}
			err = startDnsServer()
			if err != nil {
				log.Fatal(err.Error())
			}
		},
	}
)

var (
	dnsStopCmd = &cobra.Command{
		Use:   "stop",
		Short: "Stop Hoster integrated DNS Server",
		Long:  `Stop Hoster integrated DNS Server`,
		Run: func(cmd *cobra.Command, args []string) {
			err := checkInitFile()
			if err != nil {
				log.Fatal(err.Error())
			}
			err = stopDnsServer()
			if err != nil {
				log.Fatal(err.Error())
			}
		},
	}
)

var (
	dnsReloadCmd = &cobra.Command{
		Use:   "reload",
		Short: "Reload Hoster integrated DNS Server",
		Long:  `Reload Hoster integrated DNS Server`,
		Run: func(cmd *cobra.Command, args []string) {
			err := checkInitFile()
			if err != nil {
				log.Fatal(err.Error())
			}
			err = ReloadDnsServer()
			if err != nil {
				log.Fatal(err.Error())
			}
		},
	}
)

var (
	dnsShowLogCmd = &cobra.Command{
		Use:   "show-log",
		Short: "Show latest log records for the integrated DNS Server",
		Long:  `Show latest log records for the integrated DNS Server`,
		Run: func(cmd *cobra.Command, args []string) {
			err := checkInitFile()
			if err != nil {
				log.Fatal(err.Error())
			}
			err = showLogDns()
			if err != nil {
				log.Fatal(err.Error())
			}
		},
	}
)

func startDnsServer() error {
	execPath, err := os.Executable()
	if err != nil {
		return err
	}
	execFile := path.Dir(execPath) + "/dns_server"
	err = exec.Command("nohup", execFile, "&").Start()
	if err != nil {
		return err
	}
	return nil
}

func stopDnsServer() error {
	stdOut, stdErr := exec.Command("pgrep", "-lf", "dns_server").CombinedOutput()
	if stdErr != nil {
		return errors.New("DNS server is not running")
	}

	reMatch := regexp.MustCompile(`.*dns_server &.*`)
	reSplit := regexp.MustCompile(`\s+`)

	processId := ""
	for _, v := range strings.Split(string(stdOut), "\n") {
		if reMatch.MatchString(v) {
			processId = reSplit.Split(v, -1)[0]
			break
		}
	}

	_ = exec.Command("kill", "-SIGKILL", processId).Run()
	emojlog.PrintLogMessage("The process has been killed: "+processId, emojlog.Changed)

	return nil
}

func ReloadDnsServer() error {
	stdOut, stdErr := exec.Command("pgrep", "-lf", "dns_server").CombinedOutput()
	if stdErr != nil {
		return errors.New("DNS server is not running")
	}
	reMatch := regexp.MustCompile(`.*dns_server &.*`)
	reSplit := regexp.MustCompile(`\s+`)
	processId := ""
	for _, v := range strings.Split(string(stdOut), "\n") {
		if reMatch.MatchString(v) {
			processId = reSplit.Split(v, -1)[0]
			break
		}
	}
	// fmt.Println("kill", "-SIGHUP", processId)
	_ = exec.Command("kill", "-SIGHUP", processId).Run()
	emojlog.PrintLogMessage("DNS server config has been reloaded", emojlog.Changed)
	return nil
}

func showLogDns() error {
	tailCmd := exec.Command("tail", "-35", "-f", "/var/run/dns_server")
	tailCmd.Stdin = os.Stdin
	tailCmd.Stdout = os.Stdout
	tailCmd.Stderr = os.Stderr
	err := tailCmd.Run()
	if err != nil {
		return err
	}
	return nil
}
