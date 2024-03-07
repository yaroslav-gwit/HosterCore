// Copyright 2024 Hoster Authors. All rights reserved.
// Use of this source code is governed by an Apache License 2.0
// license that can be found in the LICENSE file.

package SchedulerClient

import (
	SchedulerUtils "HosterCore/internal/app/scheduler/utils"
	"encoding/json"
	"net"
)

func AddSnapshotJob(vmName string, snapshotsToKeep int, snapshotType string) error {
	c, err := net.Dial("unix", SchedulerUtils.SockAddr)
	if err != nil {
		return err
	}

	defer c.Close()

	job := SchedulerUtils.Job{}
	job.JobType = SchedulerUtils.JOB_TYPE_SNAPSHOT
	job.Snapshot.ResName = vmName
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

func AddReplicationJob(vmName string, endpoint string, key string, speedLimit int) error {
	c, err := net.Dial("unix", SchedulerUtils.SockAddr)
	if err != nil {
		return err
	}

	defer c.Close()

	job := SchedulerUtils.Job{}
	job.JobType = SchedulerUtils.JOB_TYPE_REPLICATION
	job.Replication.ResName = vmName
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

func GetJobList() (r []SchedulerUtils.Job, e error) {
	c, err := net.Dial("unix", SchedulerUtils.SockAddr)
	if err != nil {
		e = err
		return
	}
	defer c.Close()

	var job SchedulerUtils.Job
	job.JobType = SchedulerUtils.JOB_TYPE_INFO

	jsonJob, err := json.Marshal(job)
	if err != nil {
		e = err
		return
	}

	_, err = c.Write(jsonJob)
	if err != nil {
		e = err
		return
	}

	// Read the response from the socket
	var jsonResponse []byte
	buffer := make([]byte, 1024) // Adjust buffer size as needed

	for {
		n, err := c.Read(buffer)
		if err != nil {
			return nil, err
		}

		jsonResponse = append(jsonResponse, buffer[:n]...)
		if n < len(buffer) {
			break
		}
	}

	// Process the JSON response as needed
	err = json.Unmarshal(jsonResponse, &r)
	if err != nil {
		e = err
		return
	}

	return
}
