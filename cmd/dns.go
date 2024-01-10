package cmd

import (
	"HosterCore/emojlog"
	"HosterCore/osfreebsd"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path"
	"regexp"
	"strconv"
	"syscall"

	"github.com/spf13/cobra"
)

var (
	dnsCmd = &cobra.Command{
		Use:   "dns",
		Short: "Hoster integrated DNS Server",
		Long:  `Hoster integrated DNS Server`,
		Run: func(cmd *cobra.Command, args []string) {
			checkInitFile()
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
			checkInitFile()

			err := startDnsServer()
			if err != nil {
				emojlog.PrintLogMessage(err.Error(), emojlog.Error)
				os.Exit(1)
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
			checkInitFile()

			err := stopDnsServer()
			if err != nil {
				emojlog.PrintLogMessage(err.Error(), emojlog.Error)
				os.Exit(1)
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
			checkInitFile()

			err := ReloadDnsServer()
			if err != nil {
				emojlog.PrintLogMessage(err.Error(), emojlog.Error)
				os.Exit(1)
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
			checkInitFile()

			err := showLogDns()
			if err != nil {
				emojlog.PrintLogMessage(err.Error(), emojlog.Error)
				os.Exit(1)
			}
		},
	}
)

var (
	dnsStatusCmd = &cobra.Command{
		Use:   "status",
		Short: "Get DNS Server service status",
		Long:  `Get DNS Server service status.`,
		Run: func(cmd *cobra.Command, args []string) {
			checkInitFile()

			err := statusDnsServer()
			if err != nil {
				emojlog.PrintLogMessage(err.Error(), emojlog.Error)
				os.Exit(1)
			}
		},
	}
)

func startDnsServer() error {
	execPath, err := os.Executable()
	if err != nil {
		return err
	}

	os.Setenv("LOG_FILE", "/var/log/hoster_dns.log")
	os.Setenv("LOG_STDOUT", "false")

	execFile := path.Dir(execPath) + "/dns_server"
	command := exec.Command(execFile)
	command.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	err = command.Start()
	if err != nil {
		return err
	}

	return nil
}

func stopDnsServer() error {
	serviceInfo, err := dnsServerServiceInfo()
	if err != nil {
		reMatchExit1 := regexp.MustCompile(`exit status 1`)
		if reMatchExit1.MatchString(err.Error()) {
			return errors.New("DNS server is not running")
		} else {
			return err
		}
	}

	err = osfreebsd.KillProcess(osfreebsd.KillSignalTERM, serviceInfo.Pid)
	if err != nil {
		return err
	}

	emojlog.PrintLogMessage("The DNS server has been stopped using a PID: "+strconv.Itoa(serviceInfo.Pid), emojlog.Changed)
	return nil
}

func ReloadDnsServer() error {
	serviceInfo, err := dnsServerServiceInfo()
	if err != nil {
		reMatchExit1 := regexp.MustCompile(`exit status 1`)
		if reMatchExit1.MatchString(err.Error()) {
			return errors.New("DNS server is not running")
		} else {
			return err
		}
	}

	err = osfreebsd.KillProcess(osfreebsd.KillSignalHUP, serviceInfo.Pid)
	if err != nil {
		return err
	}

	emojlog.PrintLogMessage("DNS server config has been reloaded", emojlog.Changed)
	return nil
}

func showLogDns() error {
	tailCmd := exec.Command("tail", "-35", "-f", "/var/log/hoster_dns.log")
	tailCmd.Stdin = os.Stdin
	tailCmd.Stdout = os.Stdout
	tailCmd.Stderr = os.Stderr
	err := tailCmd.Run()
	if err != nil {
		return err
	}

	return nil
}

type DnsServerServiceInfoStruct struct {
	Pid     int
	Running bool
}

func dnsServerServiceInfo() (pgrepOutput DnsServerServiceInfoStruct, finalError error) {
	pids, err := osfreebsd.Pgrep("dns_server")
	if err != nil {
		reMatchExit1 := regexp.MustCompile(`exit status 1`)
		if reMatchExit1.MatchString(err.Error()) {
			finalError = errors.New("DNS server is not running")
		} else {
			finalError = err
		}
		return
	}

	if len(pids) < 1 {
		finalError = errors.New("DNS server is not running")
		return
	}

	reMatch := regexp.MustCompile(`.*dns_server$`)
	reMatchOld := regexp.MustCompile(`.*dns_server\s+&`)
	reMatchSkipLogProcess := regexp.MustCompile(`.*tail*`)

	for _, v := range pids {
		if reMatchSkipLogProcess.MatchString(v.ProcessCmd) {
			continue
		}

		if reMatch.MatchString(v.ProcessCmd) || reMatchOld.MatchString(v.ProcessCmd) {
			pgrepOutput.Pid = v.ProcessId
			pgrepOutput.Running = true
			return
		}
	}

	return
}

func statusDnsServer() error {
	dnsProcessPgrep, err := dnsServerServiceInfo()
	if err != nil {
		return err
	}

	if dnsProcessPgrep.Running {
		fmt.Println(" 🟢 DNS Server is running as PID " + strconv.Itoa(dnsProcessPgrep.Pid))
	} else {
		fmt.Println(" 🔴 DNS Server IS NOT running")
	}

	return nil
}
