package cmd

import (
	"HosterCore/emojlog"
	"HosterCore/osfreebsd"
	"encoding/json"
	"net"
	"os"

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
