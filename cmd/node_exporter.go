package cmd

import (
	"HosterCore/pkg/emojlog"
	"HosterCore/pkg/osfreebsd/fbsdkill"
	"HosterCore/pkg/osfreebsd/fbsdpgrep"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path"
	"regexp"
	"syscall"

	"github.com/spf13/cobra"
)

var (
	nodeExporterCmd = &cobra.Command{
		Use:   "node_exporter",
		Short: "Custom Node Exporter Service Control",
		Long:  `Custom Node Exporter Service Control.`,
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			checkInitFile()
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
			checkInitFile()

			err := startNodeExporter()
			if err != nil {
				emojlog.PrintLogMessage("service could not be started -> "+err.Error(), emojlog.Error)
			} else {
				emojlog.PrintLogMessage("node_exporter_custom service has been started", emojlog.Changed)
			}
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
			checkInitFile()
			err := stopNodeExporter()
			if err != nil {
				emojlog.PrintLogMessage(err.Error(), emojlog.Error)
				os.Exit(1)
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
			checkInitFile()
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

	// execFile := path.Dir(execPath) + "/node_exporter_custom"
	// err = exec.Command("nohup", execFile, "&").Start()
	// if err != nil {
	// 	return err
	// }

	execFile := path.Dir(execPath) + "/node_exporter_custom"
	command := exec.Command(execFile)
	command.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	err = command.Start()
	if err != nil {
		return err
	}

	return nil
}

func stopNodeExporter() error {
	pids := customNodeExporterServiceInfo()

	if pids.NodeExporterCustomRunning {
		err := fbsdkill.KillProcess(fbsdkill.KillSignalTERM, pids.NodeExporterCustomPid)
		if err != nil {
			return err
		}
		logOutput := fmt.Sprintf("node_exporter_custom has been stopped using the PID: %d", pids.NodeExporterCustomPid)
		emojlog.PrintLogMessage(logOutput, emojlog.Changed)
	} else {
		emojlog.PrintLogMessage("node_exporter_custom is not running", emojlog.Error)
	}

	return nil
}

func statusNodeExporter() error {
	nodeExporterPgrep := customNodeExporterServiceInfo()

	if nodeExporterPgrep.NodeExporterOfficialRunning {
		fmt.Printf(" 🟢 Node Exporter is running as PID: %d\n", nodeExporterPgrep.NodeExporterOfficialPid)
	} else {
		fmt.Println(" 🔴 Node Exporter IS NOT running")
	}

	if nodeExporterPgrep.NodeExporterCustomRunning {
		fmt.Printf(" 🟢 Node Exporter Custom is running as PID: %d\n", nodeExporterPgrep.NodeExporterCustomPid)
	} else {
		fmt.Println(" 🔴 Node Exporter Custom IS NOT running")
	}

	return nil
}

type CustomNodeExporterServiceInfo struct {
	NodeExporterOfficialRunning bool
	NodeExporterOfficialPid     int
	NodeExporterCustomRunning   bool
	NodeExporterCustomPid       int
}

func customNodeExporterServiceInfo() (pgrepOutput CustomNodeExporterServiceInfo) {
	// out, _ := exec.Command("pgrep", "-lf", "node_exporter").CombinedOutput()

	pids, err := fbsdpgrep.Pgrep("node_exporter")
	if err != nil {
		return
	}

	// reSplitSpace := regexp.MustCompile(`\s+`)
	reMatchNodeExporterOfficial := regexp.MustCompile(`/node_exporter[$|\s+]|^node_exporter[$|\s+]`)
	reMatchNodeExporterCustom := regexp.MustCompile(`node_exporter_custom`)

	for _, v := range pids {
		if reMatchNodeExporterOfficial.MatchString(v.ProcessCmd) {
			pgrepOutput.NodeExporterOfficialPid = v.ProcessId
			pgrepOutput.NodeExporterOfficialRunning = true
		}
		if reMatchNodeExporterCustom.MatchString(v.ProcessCmd) {
			pgrepOutput.NodeExporterCustomPid = v.ProcessId
			pgrepOutput.NodeExporterCustomRunning = true
		}
	}

	// for _, v := range strings.Split(string(out), "\n") {
	// 	if reMatchNodeExporterOfficial.MatchString(v) {
	// 		pid := reSplitSpace.Split(v, -1)[0]
	// 		pgrepOutput.NodeExporterOfficialPid = pid
	// 		pgrepOutput.NodeExporterOfficialRunning = true
	// 	}
	// 	if reMatchNodeExporterCustom.MatchString(v) {
	// 		pid := reSplitSpace.Split(v, -1)[0]
	// 		pgrepOutput.NodeExporterCustomPid = pid
	// 		pgrepOutput.NodeExporterCustomRunning = true
	// 	}
	// }

	return
}
