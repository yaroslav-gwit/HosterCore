// Copyright 2024 Hoster Authors. All rights reserved.
// Use of this source code is governed by an Apache License 2.0
// license that can be found in the LICENSE file.

package SchedulerUtils

type ReplicationJob struct {
	SshPort            int      `json:"ssh_port,omitempty"`
	SpeedLimit         int      `json:"speed_limit,omitempty"`
	ProgressDoneSnaps  int      `json:"done_snaps,omitempty"`
	ProgressTotalSnaps int      `json:"total_snaps,omitempty"`
	ProgressBytesDone  uint64   `json:"progress_bytes_done,omitempty"`
	ProgressBytesTotal uint64   `json:"progress_bytes_total,omitempty"`
	ZfsDataset         string   `json:"zfs_dataset,omitempty"`
	ResName            string   `json:"res_name,omitempty"`
	SshEndpoint        string   `json:"ssh_endpoint,omitempty"`
	SshKey             string   `json:"ssh_key,omitempty"`
	ScriptsRemove      []string `json:"scripts_remove"`
	ScriptsReplicate   []string `json:"scripts_replicate"`
}

type SnapshotJob struct {
	TakeImmediately bool   `json:"take_immediately,omitempty"`
	SnapshotsToKeep int    `json:"snapshots_to_keep,omitempty"`
	ZfsDataset      string `json:"zfs_dataset,omitempty"`
	ResName         string `json:"res_name,omitempty"`
	SnapshotType    string `json:"snapshot_type,omitempty"`
}

type Job struct {
	JobDone         bool           `json:"job_done,omitempty"`
	JobDoneLogged   bool           `json:"job_done_logged,omitempty"`
	JobNext         bool           `json:"job_next,omitempty"`
	JobInProgress   bool           `json:"job_in_progress,omitempty"`
	JobFailed       bool           `json:"job_failed,omitempty"`
	JobFailedLogged bool           `json:"job_failed_logged,omitempty"`
	JobId           string         `json:"job_id,omitempty"`
	JobError        string         `json:"job_error,omitempty"`
	JobType         string         `json:"job_type,omitempty"`
	Replication     ReplicationJob `json:"replication,omitempty"`
	Snapshot        SnapshotJob    `json:"snapshot,omitempty"`
}
