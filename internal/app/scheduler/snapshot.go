package main

import (
	SchedulerUtils "HosterCore/internal/app/scheduler/utils"
	zfsutils "HosterCore/internal/pkg/zfs_utils"
	"sync"
)

// Runs every 6 seconds and executes a first available job
func executeSnapshotJobs(m *sync.RWMutex) error {
	m.Lock()
	defer m.Unlock()

	for i, v := range jobs {
		if v.JobType != SchedulerUtils.JOB_TYPE_SNAPSHOT {
			continue
		}
		if v.Snapshot.ResName == replicatedVm {
			continue
		}
		if v.Snapshot.TakeImmediately {
			continue
		}

		if v.JobDone && !v.JobDoneLogged {
			logLine := "snapshot -> done for: " + v.Snapshot.ResName
			log.Info(logLine)
			jobs[i].JobDoneLogged = true
			jobs[i].JobInProgress = false
			continue
		}

		if v.JobFailed && !v.JobFailedLogged {
			logLine := "snapshot -> failed for: " + v.Snapshot.ResName
			log.Error(logLine)
			jobs[i].JobFailedLogged = true
			jobs[i].JobInProgress = false
			continue
		}

		// If the job is still in progress then break and try again during the next loop
		if v.JobInProgress {
			logLine := "snapshot -> in progress for: " + v.Snapshot.ResName
			log.Info(logLine)
			break
		}

		if !v.JobDone {
			jobs[i].JobInProgress = true
			log.Infof("snapshot -> started a new job for: %s", v.Snapshot.ResName)

			dataset, err := zfsutils.FindResourceDataset(v.Snapshot.ResName)
			if err != nil {
				log.Infof("snapshot job jailed: %v", err)
				jobs[i].JobFailed = true
				jobs[i].JobError = err.Error()
			}

			newSnap, removedSnaps, err := zfsutils.TakeScheduledSnapshot(dataset, v.Snapshot.SnapshotType, v.Snapshot.SnapshotsToKeep)
			if err != nil {
				log.Infof("snapshot job jailed: %v", err)
				jobs[i].JobFailed = true
				jobs[i].JobError = err.Error()
			} else {
				log.Infof("new snapshot taken: %s", newSnap)
				log.Infof("old snapshots removed: %v", removedSnaps)
				jobs[i].JobDone = true
			}

			break
		}
	}

	return nil
}

func executeImmediateSnapshot(m *sync.RWMutex) error {
	m.Lock()
	defer m.Unlock()

	for i, v := range jobs {
		if v.JobType != SchedulerUtils.JOB_TYPE_SNAPSHOT {
			continue
		}
		if v.Snapshot.ResName == replicatedVm {
			continue
		}
		if !v.Snapshot.TakeImmediately {
			continue
		}

		if v.JobDone && !v.JobDoneLogged {
			logLine := "immediate snapshot -> done for: " + v.Snapshot.ResName
			log.Info(logLine)
			jobs[i].JobDoneLogged = true
			jobs[i].JobInProgress = false
			continue
		}

		if v.JobFailed && !v.JobFailedLogged {
			logLine := "immediate snapshot -> failed for: " + v.Snapshot.ResName
			log.Error(logLine)
			jobs[i].JobFailedLogged = true
			jobs[i].JobInProgress = false
			continue
		}

		// If the job is still in progress then break and try again during the next loop
		if v.JobInProgress {
			logLine := "immediate snapshot -> in progress for: " + v.Snapshot.ResName
			log.Info(logLine)
			break
		}

		if !v.JobDone {
			jobs[i].JobInProgress = true
			log.Infof("immediate snapshot -> started a new job for: %s", v.Snapshot.ResName)

			dataset, err := zfsutils.FindResourceDataset(v.Snapshot.ResName)
			if err != nil {
				log.Infof("immediate snapshot job jailed: %v", err)
				jobs[i].JobFailed = true
				jobs[i].JobError = err.Error()
			}

			newSnap, removedSnaps, err := zfsutils.TakeScheduledSnapshot(dataset, v.Snapshot.SnapshotType, v.Snapshot.SnapshotsToKeep)
			if err != nil {
				log.Infof("immediate snapshot job jailed: %v", err)
				jobs[i].JobFailed = true
				jobs[i].JobError = err.Error()
			} else {
				log.Infof("new immediate snapshot taken: %s", newSnap)
				log.Infof("old immediate snapshots removed: %v", removedSnaps)
				jobs[i].JobDone = true
			}

			break
		}
	}

	return nil
}
