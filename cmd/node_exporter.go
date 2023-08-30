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
	"strings"

	"github.com/spf13/cobra"
)

var (
	nodeExporterCmd = &cobra.Command{
		Use:   "node_exporter",
		Short: "Custom Node Exporter Service Control",
		Long:  `Custom Node Exporter Service Control.`,
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
	nodeExporterStartCmd = &cobra.Command{
		Use:   "start",
		Short: "Start our custom Node Exporter",
		Long:  `Start our custom Node Exporter.`,
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			err := checkInitFile()
			if err != nil {
				log.Fatal(err.Error())
			}

			err = startNodeExporter()
			if err != nil {
				emojlog.PrintLogMessage("node_exporter_custom service could not be started", emojlog.Error)
				log.Fatal(err.Error())
			}
			emojlog.PrintLogMessage("node_exporter_custom service has been started", emojlog.Changed)
		},
	}
)

var (
	nodeExporterStopCmd = &cobra.Command{
		Use:   "stop",
		Short: "Stop our custom Node Exporter",
		Long:  `Stop our custom Node Exporter.`,
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			err := checkInitFile()
			if err != nil {
				log.Fatal(err.Error())
			}
			err = stopNodeExporter()
			if err != nil {
				log.Fatal(err.Error())
			}
		},
	}
)

var (
	nodeExporterStatusCmd = &cobra.Command{
		Use:   "status",
		Short: "Check whether the Node Exporter is running",
		Long:  `Check whether the Node Exporter is running.`,
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			err := checkInitFile()
			if err != nil {
				log.Fatal(err.Error())
			}
			statusNodeExporter()
		},
	}
)

func startNodeExporter() error {
	nodeExporterPgrep := customNodeExporterServiceInfo()
	if nodeExporterPgrep.NodeExporterCustomRunning {
		return errors.New("node_exporter_custom is already running")
	}
	if !nodeExporterPgrep.NodeExporterOfficialRunning {
		return errors.New("node_exporter service is not running, please make sure you start it before using our custom exporter")
	}

	execPath, err := os.Executable()
	if err != nil {
		return err
	}

	execFile := path.Dir(execPath) + "/node_exporter_custom"
	err = exec.Command("nohup", execFile, "&").Start()
	if err != nil {
		return err
	}

	return nil
}

func stopNodeExporter() error {
	nodeExporterPgrep := customNodeExporterServiceInfo()

	if nodeExporterPgrep.NodeExporterCustomRunning {
		_ = exec.Command("kill", "-SIGTERM", nodeExporterPgrep.NodeExporterCustomPid).Run()
		emojlog.PrintLogMessage("node_exporter_custom was stopped using PID: "+nodeExporterPgrep.NodeExporterCustomPid, emojlog.Changed)
	} else {
		emojlog.PrintLogMessage("node_exporter_custom is not running"+nodeExporterPgrep.NodeExporterCustomPid, emojlog.Error)
	}

	return nil
}

func statusNodeExporter() error {
	nodeExporterPgrep := customNodeExporterServiceInfo()

	if nodeExporterPgrep.NodeExporterOfficialRunning {
		fmt.Println(" 🟢 Node Exporter is running as " + nodeExporterPgrep.NodeExporterOfficialPid)
	} else {
		fmt.Println(" 🔴 Node Exporter IS NOT running")
	}

	if nodeExporterPgrep.NodeExporterCustomRunning {
		fmt.Println(" 🟢 Custom Node Exporter is running as " + nodeExporterPgrep.NodeExporterCustomPid)
	} else {
		fmt.Println(" 🔴 Custom Node Exporter IS NOT running")
	}

	return nil
}

type CustomNodeExporterServiceInfo struct {
	NodeExporterOfficialRunning bool
	NodeExporterOfficialPid     string
	NodeExporterCustomRunning   bool
	NodeExporterCustomPid       string
}

func customNodeExporterServiceInfo() (pgrepOutput CustomNodeExporterServiceInfo) {
	out, _ := exec.Command("pgrep", "-lf", "node_exporter").CombinedOutput()

	reSplitSpace := regexp.MustCompile(`\s+`)
	reMatchNodeExporterOfficial := regexp.MustCompile(`.*node_exporter[$|\s+]`)
	reMatchNodeExporterCustom := regexp.MustCompile(`.*node_exporter_custom`)

	for _, v := range strings.Split(string(out), "\n") {
		if reMatchNodeExporterOfficial.MatchString(v) {
			pid := reSplitSpace.Split(v, -1)[0]
			pgrepOutput.NodeExporterOfficialPid = pid
			pgrepOutput.NodeExporterOfficialRunning = true
		}
		if reMatchNodeExporterCustom.MatchString(v) {
			pid := reSplitSpace.Split(v, -1)[0]
			pgrepOutput.NodeExporterCustomPid = pid
			pgrepOutput.NodeExporterCustomRunning = true
		}
	}

	return
}
