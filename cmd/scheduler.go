//go:build freebsd
// +build freebsd

package cmd

import (
	SchedulerClient "HosterCore/internal/app/scheduler/client"
	SchedulerUtils "HosterCore/internal/app/scheduler/utils"
	"HosterCore/internal/pkg/emojlog"
	FreeBSDKill "HosterCore/internal/pkg/freebsd/kill"
	FreeBSDPgrep "HosterCore/internal/pkg/freebsd/pgrep"
	HosterCliJson "HosterCore/internal/pkg/hoster/cli_json"
	HosterTables "HosterCore/internal/pkg/hoster/cli_tables"
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
	schedulerCmd = &cobra.Command{
		Use:   "scheduler",
		Short: "Hoster Scheduling Service",
		Long:  `Hoster Scheduling Service.`,
		Run: func(cmd *cobra.Command, args []string) {
			checkInitFile()
			cmd.Help()
		},
	}
)

var (
	schedulerStartCmd = &cobra.Command{
		Use:   "start",
		Short: "Start a background Scheduler service",
		Long:  `Start a background Scheduler service.`,
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			checkInitFile()

			err := startSchedulerService()
			if err != nil {
				emojlog.PrintLogMessage(err.Error(), emojlog.Error)
				os.Exit(1)
			}
		},
	}
)

func startSchedulerService() error {
	execPath, err := os.Executable()
	if err != nil {
		return err
	}
	os.Setenv("LOG_FILE", "/var/log/hoster_scheduler.log")

	execFile := path.Dir(execPath) + "/scheduler"
	command := exec.Command(execFile)
	command.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	err = command.Start()
	if err != nil {
		return err
	}

	return nil
}

var (
	schedulerStatusCmd = &cobra.Command{
		Use:   "status",
		Short: "Show a status for the background Scheduler service",
		Long:  "Show a status for the background Scheduler service.",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			checkInitFile()

			err := statusSchedulerService()
			if err != nil {
				emojlog.PrintLogMessage(err.Error(), emojlog.Error)
				os.Exit(1)
			}
		},
	}
)

var (
	// schedulerListJson       bool
	// schedulerInfoJobID      string
	schedulerInfoJsonPretty bool

	schedulerInfoCmd = &cobra.Command{
		Use:   "info [jobID]",
		Short: "Show info for one of the scheduled jobs",
		Long:  "Show a JSON-formatted info for one of the scheduled jobs",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			checkInitFile()

			err := HosterCliJson.GenerateSchedulerJobInfo(args[0], schedulerInfoJsonPretty)
			if err != nil {
				emojlog.PrintLogMessage(err.Error(), emojlog.Error)
				os.Exit(1)
			}
		},
	}
)

var (
	schedulerListUnix       bool
	schedulerListJson       bool
	schedulerListJsonPretty bool

	schedulerListCmd = &cobra.Command{
		Use:   "list",
		Short: "Show a list of scheduled jobs",
		Long:  "Show a list of scheduled, completed, and in-progress jobs.",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			checkInitFile()

			if schedulerListJson || schedulerListJsonPretty {
				err := HosterCliJson.GenerateSchedulerJson(schedulerListJsonPretty)
				if err != nil {
					emojlog.PrintLogMessage(err.Error(), emojlog.Error)
					os.Exit(1)
				}
			} else {
				err := HosterTables.GenerateJobsTable(schedulerListUnix)
				if err != nil {
					emojlog.PrintLogMessage(err.Error(), emojlog.Error)
					os.Exit(1)
				}
			}
		},
	}
)

func statusSchedulerService() error {
	pids, err := FreeBSDPgrep.Pgrep("scheduler")
	if err != nil {
		fmt.Println(" 🔴 Scheduler IS NOT running")
		return nil
	}

	if len(pids) < 1 {
		fmt.Println(" 🔴 Scheduler IS NOT running")
		return nil
	}

	reMatchScheduler := regexp.MustCompile(`/scheduler`)
	for _, v := range pids {
		if reMatchScheduler.MatchString(v.ProcessCmd) {
			fmt.Println(" 🟢 Scheduler is running as PID " + strconv.Itoa(v.ProcessId))

			// fmt.Println()
			// resp, err := SchedulerClient.GetJobList()
			// if err != nil {
			// 	fmt.Println("ERROR: " + err.Error())
			// }

			// out, err := json.MarshalIndent(resp, "", "   ")
			// if err != nil {
			// 	fmt.Println("ERROR: " + err.Error())
			// }
			// fmt.Println(string(out))

			return nil
		}
	}

	fmt.Println(" 🔴 Scheduler IS NOT running")
	return nil
}

var (
	schedulerStopCmd = &cobra.Command{
		Use:   "stop",
		Short: "Stop the background Scheduler service",
		Long:  "Stop the background Scheduler service.",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			checkInitFile()

			err := stopSchedulerService()
			if err != nil {
				emojlog.PrintLogMessage(err.Error(), emojlog.Error)
				os.Exit(1)
			}
		},
	}
)

func stopSchedulerService() error {
	pids, err := FreeBSDPgrep.Pgrep("scheduler")
	if err != nil {
		emojlog.PrintLogMessage("Scheduler is not running", emojlog.Error)
		return nil
	}

	if len(pids) < 1 {
		emojlog.PrintLogMessage("Scheduler is not running", emojlog.Error)
		return nil
	}

	reMatchScheduler := regexp.MustCompile(`/scheduler`)
	for _, v := range pids {
		if reMatchScheduler.MatchString(v.ProcessCmd) {
			err := FreeBSDKill.KillProcess(FreeBSDKill.KillSignalTERM, v.ProcessId)
			if err != nil {
				emojlog.PrintLogMessage("Could not stop the Scheduler "+err.Error(), emojlog.Error)
				return nil
			}
			emojlog.PrintLogMessage("Scheduler has been stopped using a PID "+fmt.Sprintf("%d", v.ProcessId), emojlog.Changed)
			return nil
		}
	}

	emojlog.PrintLogMessage("Scheduler is not running", emojlog.Error)
	return nil
}

var (
	schedulerReplicateKey        string
	schedulerReplicateEndpoint   string
	schedulerReplicatePort       int
	schedulerReplicateSpeedLimit int

	schedulerReplicateCmd = &cobra.Command{
		Use:   "replicate [VM or Jail name]",
		Short: "Use the Scheduling Service to start the resource replication",
		Long:  `Use the Scheduling Service to start the resource replication in the background mode.`,
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			checkInitFile()

			job := SchedulerUtils.ReplicationJob{}
			job.ResName = args[0]
			job.SshKey = schedulerReplicateKey
			job.SshEndpoint = schedulerReplicateEndpoint
			job.SshPort = schedulerReplicatePort
			job.SpeedLimit = schedulerReplicateSpeedLimit

			// err := SchedulerClient.Replicate(job)
			err := SchedulerClient.AddReplicationJob(job)
			if err != nil {
				emojlog.PrintLogMessage(err.Error(), emojlog.Error)
				os.Exit(1)
			}

			emojlog.PrintLogMessage("A new background replication job has been added for "+args[0], emojlog.Changed)
		},
	}
)

var (
	schedulerReplicateByTagKey        string
	schedulerReplicateByTagEndpoint   string
	schedulerReplicateByTagPort       int
	schedulerReplicateByTagSpeedLimit int

	schedulerReplicateByTagCmd = &cobra.Command{
		Use:   "replicate-by-tag [tag]",
		Short: "Use the Scheduling Service to start multi-resource replication",
		Long:  `Use the Scheduling Service to start multi-resource replication in the background mode.`,
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			checkInitFile()

			err := SchedulerClient.AddReplicationByTagJob(args[0], schedulerReplicateByTagKey, schedulerReplicateByTagEndpoint, schedulerReplicateByTagPort, schedulerReplicateByTagSpeedLimit)
			if err != nil {
				emojlog.PrintErrorMessage(err.Error())
				os.Exit(1)
			}

			emojlog.PrintChangedMessage("Replication jobs have been added for a tag: " + args[0])
		},
	}
)

var (
	schedulerSnapshotToKeep int
	schedulerSnapshotType   string

	schedulerSnapshotCmd = &cobra.Command{
		Use:   "snapshot [VM or Jail name]",
		Short: "Use the Scheduling Service to snapshot a VM or a Jail",
		Long:  `Use the Scheduling Service to snapshot a VM or a Jail in the background mode.`,
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			checkInitFile()

			_, err := SchedulerClient.AddSnapshotJob(args[0], schedulerSnapshotToKeep, schedulerSnapshotType, false)
			if err != nil {
				emojlog.PrintLogMessage(err.Error(), emojlog.Error)
				os.Exit(1)
			}

			emojlog.PrintLogMessage("A new background snapshot job has been added for "+args[0], emojlog.Changed)
		},
	}
)

var (
	schedulerSnapshotAllToKeep int
	schedulerSnapshotAllType   string

	schedulerSnapshotAllCmd = &cobra.Command{
		Use:   "snapshot-all",
		Short: "Use the Scheduling Service to snapshot all VMs and Jails",
		Long:  `Use the Scheduling Service to snapshot all VMs and Jails in the background mode.`,
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			checkInitFile()

			err := SchedulerClient.AddSnapshotAllJob(schedulerSnapshotAllToKeep, schedulerSnapshotAllType)
			if err != nil {
				emojlog.PrintLogMessage(err.Error(), emojlog.Error)
				os.Exit(1)
			}

			// vms, err := HosterVmUtils.ListJsonApi()
			// if err != nil {
			// 	emojlog.PrintLogMessage("could not get a list of VMs: "+err.Error(), emojlog.Error)
			// 	os.Exit(1)
			// }
			// for _, v := range vms {
			// 	if !v.Running {
			// 		continue
			// 	}
			// 	if v.Backup {
			// 		continue
			// 	}

			// 	err := SchedulerClient.AddSnapshotJob(v.Name, schedulerSnapshotAllToKeep, schedulerSnapshotAllType, false)
			// 	if err != nil {
			// 		emojlog.PrintLogMessage(err.Error(), emojlog.Error)
			// 	} else {
			// 		emojlog.PrintLogMessage("A new background snapshot job has been added for a VM: "+v.Name, emojlog.Changed)
			// 	}
			// }

			// jails, err := HosterJailUtils.ListJsonApi()
			// if err != nil {
			// 	emojlog.PrintLogMessage("could not get a list of Jails: "+err.Error(), emojlog.Error)
			// 	os.Exit(1)
			// }
			// for _, v := range jails {
			// 	if !v.Running {
			// 		continue
			// 	}
			// 	if v.Backup {
			// 		continue
			// 	}

			// 	err = SchedulerClient.AddSnapshotJob(v.Name, schedulerSnapshotAllToKeep, schedulerSnapshotAllType, false)
			// 	if err != nil {
			// 		emojlog.PrintLogMessage(err.Error(), emojlog.Error)
			// 	} else {
			// 		emojlog.PrintLogMessage("A new background snapshot job has been added for a Jail: "+v.Name, emojlog.Changed)
			// 	}
			// }
		},
	}
)

var (
	schedulerShowLogCmd = &cobra.Command{
		Use:   "show-log",
		Short: "Show latest log records for the Scheduler service",
		Long:  `Show latest log records for the Scheduler service.`,
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			checkInitFile()

			err := showLogScheduler()
			if err != nil {
				emojlog.PrintLogMessage(err.Error(), emojlog.Error)
				os.Exit(1)
			}
		},
	}
)

func showLogScheduler() error {
	tailCmd := exec.Command("tail", "-35", "-f", "/var/log/hoster_scheduler.log")

	tailCmd.Stdin = os.Stdin
	tailCmd.Stdout = os.Stdout
	tailCmd.Stderr = os.Stderr

	err := tailCmd.Run()
	if err != nil {
		return err
	}

	return nil
}
