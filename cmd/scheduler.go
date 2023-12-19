package cmd

import (
	"HosterCore/emojlog"
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
	schedulerReplicateVmName     string
	schedulerReplicateEndpoint   string
	schedulerReplicateKey        string
	schedulerReplicateSpeedLimit int

	schedulerReplicateCmd = &cobra.Command{
		Use:   "replicate",
		Short: "Use the Scheduling Service to start the VM replication",
		Long:  `Use the Scheduling Service to start the VM replication in the background mode.`,
		Run: func(cmd *cobra.Command, args []string) {
			checkInitFile()

			err := addReplicationJob(schedulerReplicateVmName, schedulerReplicateEndpoint, schedulerReplicateKey, schedulerReplicateSpeedLimit)
			if err != nil {
				emojlog.PrintLogMessage(err.Error(), emojlog.Error)
				os.Exit(1)
			}

			emojlog.PrintLogMessage("A new background replication job has been added for "+schedulerReplicateVmName, emojlog.Changed)
		},
	}
)

var (
	schedulerSnapshotVmName string
	schedulerSnapshotType   string
	schedulerSnapshotToKeep int

	schedulerSnapshotCmd = &cobra.Command{
		Use:   "snapshot",
		Short: "Use the Scheduling Service to snapshot the VM",
		Long:  `Use the Scheduling Service to snapshot the VM in the background mode.`,
		Run: func(cmd *cobra.Command, args []string) {
			checkInitFile()
			cmd.Help()
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
