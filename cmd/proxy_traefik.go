package cmd

import (
	"errors"
	"fmt"
	"hoster/emojlog"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/spf13/cobra"
)

var (
	proxyTraefikCmd = &cobra.Command{
		Use:   "traefik",
		Short: "Minimalistic Traefik integration",
		Long:  `Minimalistic Traefik integration.`,
		Args:  cobra.NoArgs,
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
	proxyTraefikStartCmd = &cobra.Command{
		Use:   "start",
		Short: "Start Traefik reverse proxy service",
		Long:  `Start Traefik reverse proxy service.`,
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			err := checkInitFile()
			if err != nil {
				log.Fatal(err.Error())
			}

			err = startTraefik()
			if err != nil {
				emojlog.PrintLogMessage(err.Error(), emojlog.Error)
				os.Exit(1)
			}
			emojlog.PrintLogMessage("Traefik service has been started", emojlog.Changed)
		},
	}
)

var (
	proxyTraefikStopCmd = &cobra.Command{
		Use:   "stop",
		Short: "Stop Traefik reverse proxy service",
		Long:  `Stop Traefik reverse proxy service.`,
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			err := checkInitFile()
			if err != nil {
				log.Fatal(err.Error())
			}

			err = stopTraefik()
			if err != nil {
				log.Fatal(err.Error())
			}
		},
	}
)

var (
	proxyTraefikStatusCmd = &cobra.Command{
		Use:   "status",
		Short: "Traefik reverse proxy service status",
		Long:  `Traefik reverse proxy service status.`,
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			err := checkInitFile()
			if err != nil {
				log.Fatal(err.Error())
			}

			err = statusTraefik()
			if err != nil {
				log.Fatal(err.Error())
			}
		},
	}
)

func startTraefik() error {
	traefikPgrep := traefikServiceInfo()

	if traefikPgrep.ProcessIsRunning {
		return errors.New("service Traefik is already running")
	}

	err := exec.Command("nohup", "/opt/traefik/traefik", "--configFile=/opt/traefik/traefik.yaml", "&").Start()
	if err != nil {
		return err
	}

	return nil
}

func stopTraefik() error {
	traefikPgrep := traefikServiceInfo()

	if traefikPgrep.ProcessIsRunning {
		_ = exec.Command("kill", "-SIGTERM", traefikPgrep.ProcessPid).Run()
		emojlog.PrintLogMessage("Traefik was stopped using PID: "+traefikPgrep.ProcessPid, emojlog.Changed)
	} else {
		emojlog.PrintLogMessage("Traefik is not running", emojlog.Error)
	}

	return nil
}

func statusTraefik() error {
	traefikPgrep := traefikServiceInfo()

	if traefikPgrep.ProcessIsRunning {
		fmt.Println(" 🟢 Traefik is running as " + traefikPgrep.ProcessPid)
		fmt.Println("    to view live service logs execute: tail -f /opt/traefik/log_service.log")
		fmt.Println("    to view live access logs execute:  tail -f /opt/traefik/log_access.log")
	} else {
		fmt.Println(" 🔴 Traefik IS NOT running")
	}

	return nil
}

type TraefikServiceInfoStruct struct {
	ProcessIsRunning bool
	ProcessPid       string
}

func traefikServiceInfo() (pgrepOutput TraefikServiceInfoStruct) {
	out, _ := exec.Command("pgrep", "-lf", "node_exporter").CombinedOutput()

	reSplitSpace := regexp.MustCompile(`\s+`)
	reMatchTraefik := regexp.MustCompile(`.*/opt/traefik/traefik.*`)

	for _, v := range strings.Split(string(out), "\n") {
		if reMatchTraefik.MatchString(v) {
			pid := reSplitSpace.Split(v, -1)[0]
			pgrepOutput.ProcessPid = pid
			pgrepOutput.ProcessIsRunning = true
		}
	}

	return
}