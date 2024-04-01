// Copyright 2024 Hoster Authors. All rights reserved.
// Use of this source code is governed by an Apache License 2.0
// license that can be found in the LICENSE file.

package SchedulerClient

import (
	SchedulerUtils "HosterCore/internal/app/scheduler/utils"
	HosterJailUtils "HosterCore/internal/pkg/hoster/jail/utils"
	HosterVmUtils "HosterCore/internal/pkg/hoster/vm/utils"
	"encoding/json"
	"fmt"
	"net"
)

func AddSnapshotJob(resName string, snapshotsToKeep int, snapshotType string, takeImmediately bool) error {
	// Res found check
	resFound := false

	if !resFound {
		jails, err := HosterJailUtils.ListAllSimple()
		if err != nil {
			return err
		}
		for i := range jails {
			if jails[i].JailName == resName {
				resFound = true
			}
		}
	}

	if !resFound {
		vms, err := HosterVmUtils.ListAllSimple()
		if err != nil {
			return err
		}
		for i := range vms {
			if vms[i].VmName == resName {
				resFound = true
			}
		}
	}

	if !resFound {
		return fmt.Errorf("resource was not found")
	}
	// EOF Res found check

	c, err := net.Dial("unix", SchedulerUtils.SockAddr)
	if err != nil {
		return err
	}
	defer c.Close()

	job := SchedulerUtils.Job{}
	job.JobType = SchedulerUtils.JOB_TYPE_SNAPSHOT
	job.Snapshot.SnapshotsToKeep = snapshotsToKeep
	job.Snapshot.TakeImmediately = takeImmediately
	job.Snapshot.SnapshotType = snapshotType
	job.Snapshot.ResName = resName

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

func AddSnapshotAllJob(snapshotsToKeep int, snapshotType string) error {
	// Get all running resoruces
	jails, err := HosterJailUtils.ListJsonApi()
	if err != nil {
		return err
	}
	vms, err := HosterVmUtils.ListJsonApi()
	if err != nil {
		return err
	}
	// Get all running resoruces

	job := SchedulerUtils.Job{}
	job.JobType = SchedulerUtils.JOB_TYPE_SNAPSHOT
	job.Snapshot.SnapshotsToKeep = snapshotsToKeep
	job.Snapshot.SnapshotType = snapshotType

	for i := range vms {
		if vms[i].Backup || !vms[i].Running {
			continue
		}

		c, err := net.Dial("unix", SchedulerUtils.SockAddr)
		if err != nil {
			c.Close()
			return err
		}

		job.Snapshot.ResName = vms[i].Name
		jsonJob, err := json.Marshal(job)
		if err != nil {
			c.Close()
			return err
		}

		_, err = c.Write(jsonJob)
		if err != nil {
			c.Close()
			return err
		}

		c.Close()
	}

	for i := range jails {
		if jails[i].Backup || !jails[i].Running {
			continue
		}

		c, err := net.Dial("unix", SchedulerUtils.SockAddr)
		if err != nil {
			c.Close()
			return err
		}

		job.Snapshot.ResName = jails[i].Name
		jsonJob, err := json.Marshal(job)
		if err != nil {
			c.Close()
			return err
		}

		_, err = c.Write(jsonJob)
		if err != nil {
			c.Close()
			return err
		}
	}

	return nil
}

func AddSnapshotDestroyJob(resName string, snapshotName string) error {
	// Res found check
	resFound := false

	if !resFound {
		jails, err := HosterJailUtils.ListAllSimple()
		if err != nil {
			return err
		}
		for i := range jails {
			if jails[i].JailName == resName {
				resFound = true
			}
		}
	}

	if !resFound {
		vms, err := HosterVmUtils.ListAllSimple()
		if err != nil {
			return err
		}
		for i := range vms {
			if vms[i].VmName == resName {
				resFound = true
			}
		}
	}

	if !resFound {
		return fmt.Errorf("resource was not found")
	}
	// EOF Res found check

	c, err := net.Dial("unix", SchedulerUtils.SockAddr)
	if err != nil {
		return err
	}
	defer c.Close()

	job := SchedulerUtils.Job{}
	job.JobType = SchedulerUtils.JOB_TYPE_SNAPSHOT_DESTROY
	job.Snapshot.SnapshotName = snapshotName
	job.Snapshot.TakeImmediately = true
	job.Snapshot.ResName = resName

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

func AddSnapshotRollbackJob(resName string, snapshotName string) error {
	// Res found check
	resFound := false

	if !resFound {
		jails, err := HosterJailUtils.ListAllSimple()
		if err != nil {
			return err
		}
		for i := range jails {
			if jails[i].JailName == resName {
				resFound = true
			}
		}
	}

	if !resFound {
		vms, err := HosterVmUtils.ListAllSimple()
		if err != nil {
			return err
		}
		for i := range vms {
			if vms[i].VmName == resName {
				resFound = true
			}
		}
	}

	if !resFound {
		return fmt.Errorf("resource was not found")
	}
	// EOF Res found check

	c, err := net.Dial("unix", SchedulerUtils.SockAddr)
	if err != nil {
		return err
	}
	defer c.Close()

	job := SchedulerUtils.Job{}
	job.JobType = SchedulerUtils.JOB_TYPE_SNAPSHOT_ROLLBACK
	job.Snapshot.SnapshotName = snapshotName
	job.Snapshot.TakeImmediately = true
	job.Snapshot.ResName = resName

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
