// Copyright 2024 Hoster Authors. All rights reserved.
// Use of this source code is governed by an Apache License 2.0
// license that can be found in the LICENSE file.

package SchedulerClient

import (
	SchedulerUtils "HosterCore/internal/app/scheduler/utils"
	"encoding/json"
	"net"
)

func AddSnapshotJob(vmName string, snapshotsToKeep int, snapshotType string, takeImmediately bool) error {
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
	job.Snapshot.TakeImmediately = takeImmediately

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
