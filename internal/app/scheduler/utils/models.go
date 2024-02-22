// Copyright 2024 Hoster Authors. All rights reserved.
// Use of this source code is governed by an Apache License 2.0
// license that can be found in the LICENSE file.

package SchedulerUtils

type ReplicationJob struct {
	ZfsDataset         string `json:"zfs_dataset,omitempty"`
	VmName             string `json:"vm_name,omitempty"`
	SshEndpoint        string `json:"ssh_endpoint,omitempty"`
	SshKey             string `json:"ssh_key,omitempty"`
	BufferSpeedLimit   int    `json:"speed_limit,omitempty"`
	ProgressBytes      int    `json:"progress_bytes,omitempty"`
	ProgressPercent    int    `json:"progress_percent,omitempty"`
	ProgressDoneSnaps  int    `json:"done_snaps,omitempty"`
	ProgressTotalSnaps int    `json:"total_snaps,omitempty"`
}

type SnapshotJob struct {
	ZfsDataset      string `json:"zfs_dataset,omitempty"`
	VmName          string `json:"vm_name,omitempty"`
	SnapshotsToKeep int    `json:"snapshots_to_keep,omitempty"`
	SnapshotType    string `json:"snapshot_type,omitempty"`
}

type Job struct {
	JobDone         bool           `json:"job_done,omitempty"`
	JobId           string         `json:"job_id,omitempty"`
	JobDoneLogged   bool           `json:"job_done_logged,omitempty"`
	JobNext         bool           `json:"job_next,omitempty"`
	JobInProgress   bool           `json:"job_in_progress,omitempty"`
	JobFailed       bool           `json:"job_failed,omitempty"`
	JobFailedLogged bool           `json:"job_failed_logged,omitempty"`
	JobError        string         `json:"job_error,omitempty"`
	JobType         string         `json:"job_type,omitempty"`
	Replication     ReplicationJob `json:"replication,omitempty"`
	Snapshot        SnapshotJob    `json:"snapshot,omitempty"`
}
