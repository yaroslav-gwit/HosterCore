package cmd

import (
	"HosterCore/emojlog"
	"HosterCore/osfreebsd"
	"encoding/json"
	"fmt"
	"net"
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
	os.Setenv("LOG_STDOUT", "false")
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

func statusSchedulerService() error {
	pids, err := osfreebsd.Pgrep("scheduler")
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
	pids, err := osfreebsd.Pgrep("scheduler")
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
			err := osfreebsd.KillProcess(osfreebsd.KillSignalTERM, v.ProcessId)
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
	schedulerReplicateEndpoint   string
	schedulerReplicateKey        string
	schedulerReplicateSpeedLimit int

	schedulerReplicateCmd = &cobra.Command{
		Use:   "replicate [VM or Jail name]",
		Short: "Use the Scheduling Service to start the VM replication",
		Long:  `Use the Scheduling Service to start the VM replication in the background mode.`,
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			checkInitFile()

			err := addReplicationJob(args[0], schedulerReplicateEndpoint, schedulerReplicateKey, schedulerReplicateSpeedLimit)
			if err != nil {
				emojlog.PrintLogMessage(err.Error(), emojlog.Error)
				os.Exit(1)
			}

			emojlog.PrintLogMessage("A new background replication job has been added for "+args[0], emojlog.Changed)
		},
	}
)

var (
	schedulerSnapshotToKeep int
	schedulerSnapshotType   string

	schedulerSnapshotCmd = &cobra.Command{
		Use:   "snapshot [VM or Jail name]",
		Short: "Use the Scheduling Service to snapshot the VM",
		Long:  `Use the Scheduling Service to snapshot the VM in the background mode.`,
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			checkInitFile()

			err := addSnapshotJob(args[0], schedulerSnapshotToKeep, schedulerSnapshotType)
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

			hostname, err := osfreebsd.SysctlKernHostname()
			if err != nil {
				emojlog.PrintLogMessage("could not get a hostname: "+err.Error(), emojlog.Error)
				os.Exit(1)
			}

			for _, v := range getAllVms() {
				if !VmLiveCheck(v) {
					continue
				}

				vmConf := vmConfig(v)
				if vmConf.ParentHost != hostname {
					continue
				}

				err := addSnapshotJob(v, schedulerSnapshotAllToKeep, schedulerSnapshotAllType)
				if err != nil {
					emojlog.PrintLogMessage(err.Error(), emojlog.Error)
				} else {
					emojlog.PrintLogMessage("A new background snapshot job has been added for a VM: "+v, emojlog.Changed)
				}
			}

			jailList, err := GetAllJailsList()
			if err != nil {
				emojlog.PrintLogMessage("could not get a list of Jails: "+err.Error(), emojlog.Error)
				os.Exit(1)
			}
			for _, v := range jailList {
				jailConf, err := GetJailConfig(v, true)
				if err != nil {
					continue
				}
				if jailConf.Parent != hostname {
					continue
				}

				err = addSnapshotJob(v, schedulerSnapshotAllToKeep, schedulerSnapshotAllType)
				if err != nil {
					emojlog.PrintLogMessage(err.Error(), emojlog.Error)
				} else {
					emojlog.PrintLogMessage("A new background snapshot job has been added for a Jail: "+v, emojlog.Changed)
				}
			}
		},
	}
)

// Hardcoded code copy, to avoid circular imports
// Will eliminate it at some point, after the refactoring is complete
// And when it will be possible to import without the circular import issues
var SockAddr = "/var/run/hoster_scheduler.sock"

const (
	JOB_TYPE_REPLICATION = "replication"
	JOB_TYPE_SNAPSHOT    = "snapshot"
)

type ReplicationJob struct {
	// ZfsDataset       string `json:"zfs_dataset"`
	VmName           string `json:"vm_name"`
	SshEndpoint      string `json:"ssh_endpoint"`
	SshKey           string `json:"ssh_key"`
	BufferSpeedLimit int    `json:"speed_limit"`
	ProgressBytes    int    `json:"progress_bytes"`
	ProgressPercent  int    `json:"progress_percent"`
}

type SnapshotJob struct {
	// ZfsDataset      string `json:"zfs_dataset"`
	VmName          string `json:"vm_name"`
	SnapshotsToKeep int    `json:"snapshots_to_keep"`
	SnapshotType    string `json:"snapshot_type"`
}

type Job struct {
	JobDone       bool           `json:"job_done"`
	JobNext       bool           `json:"job_next"`
	JobInProgress bool           `json:"job_in_progress"`
	JobFailed     bool           `json:"job_failed"`
	JobError      string         `json:"job_error"`
	JobType       string         `json:"job_type"`
	Replication   ReplicationJob `json:"replication"`
	Snapshot      SnapshotJob    `json:"snapshot"`
}

func addReplicationJob(vmName string, endpoint string, key string, speedLimit int) error {
	c, err := net.Dial("unix", SockAddr)
	if err != nil {
		return err
	}

	defer c.Close()

	job := Job{}
	job.JobType = JOB_TYPE_REPLICATION
	job.Replication.VmName = vmName
	job.Replication.SshEndpoint = endpoint
	job.Replication.SshKey = key
	job.Replication.BufferSpeedLimit = speedLimit

	jsonJob, err := json.Marshal(job)
	if err != nil {
		return err
	}

	_, err = c.Write(jsonJob)
	if err != nil {
		return err
	}

	return nil
}

func addSnapshotJob(vmName string, snapshotsToKeep int, snapshotType string) error {
	c, err := net.Dial("unix", SockAddr)
	if err != nil {
		return err
	}

	defer c.Close()

	job := Job{}
	job.JobType = JOB_TYPE_SNAPSHOT
	job.Snapshot.VmName = vmName
	job.Snapshot.SnapshotType = snapshotType
	job.Snapshot.SnapshotsToKeep = snapshotsToKeep

	jsonJob, err := json.Marshal(job)
	if err != nil {
		return err
	}

	_, err = c.Write(jsonJob)
	if err != nil {
		return err
	}

	return nil
}

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
