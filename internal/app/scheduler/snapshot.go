package main

import (
	SchedulerUtils "HosterCore/internal/app/scheduler/utils"
	HosterJail "HosterCore/internal/pkg/hoster/jail"
	HosterVm "HosterCore/internal/pkg/hoster/vm"
	HosterVmUtils "HosterCore/internal/pkg/hoster/vm/utils"
	zfsutils "HosterCore/internal/pkg/zfs_utils"
	"strings"
	"sync"
	"time"
)

// Runs every 6 seconds and executes a first available job
func executeSnapshotJobs(m *sync.RWMutex) error {
	m.Lock()
	defer m.Unlock()

	for i, v := range jobs {
		if v.JobType != SchedulerUtils.JOB_TYPE_SNAPSHOT {
			continue
		}
		if v.Snapshot.ResName == getReplicatedVm() {
			continue
		}
		// if v.Replication.ResName == snapShottedVM {
		// 	continue
		// }
		if snapshotMap[v.Snapshot.ResName] {
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

			// snapShottedVM = jobs[i].Snapshot.ResName
			snapshotMap[jobs[i].Snapshot.ResName] = true
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
			snapshotMap[jobs[i].Snapshot.ResName] = false
			jobs[i].TimeFinished = time.Now().Unix()

			break
		}
	}

	return nil
}

func executeImmediateSnapshot(m *sync.RWMutex) error {
	m.Lock()
	defer m.Unlock()

IMMEDIATE_SNAPSHOT:
	for i, v := range jobs {
		if v.JobType == SchedulerUtils.JOB_TYPE_SNAPSHOT || v.JobType == SchedulerUtils.JOB_TYPE_SNAPSHOT_DESTROY || v.JobType == SchedulerUtils.JOB_TYPE_SNAPSHOT_ROLLBACK {
			_ = 0
		} else {
			continue IMMEDIATE_SNAPSHOT
		}
		if v.Snapshot.ResName == getReplicatedVm() {
			continue IMMEDIATE_SNAPSHOT
		}
		// if v.Replication.ResName == snapShottedVM {
		// 	continue
		// }
		if snapshotMap[v.Snapshot.ResName] {
			continue IMMEDIATE_SNAPSHOT
		}
		if !v.Snapshot.TakeImmediately {
			continue IMMEDIATE_SNAPSHOT
		}

		if v.JobDone && !v.JobDoneLogged {
			logLine := "immediate snapshot -> done for: " + v.Snapshot.ResName
			log.Info(logLine)
			jobs[i].JobDoneLogged = true
			jobs[i].JobInProgress = false
			continue IMMEDIATE_SNAPSHOT
		}

		if v.JobFailed && !v.JobFailedLogged {
			logLine := "immediate snapshot -> failed for: " + v.Snapshot.ResName
			log.Error(logLine)
			jobs[i].JobFailedLogged = true
			jobs[i].JobInProgress = false
			continue IMMEDIATE_SNAPSHOT
		}

		// If the job is still in progress then break and try again during the next loop
		if v.JobInProgress {
			logLine := "immediate snapshot -> in progress for: " + v.Snapshot.ResName
			log.Info(logLine)
			break IMMEDIATE_SNAPSHOT
		}

		if !v.JobDone {
			jobs[i].JobInProgress = true
			log.Infof("immediate snapshot -> started a new job for: %s", v.Snapshot.ResName)

			dataset, err := zfsutils.FindResourceDataset(v.Snapshot.ResName)
			if err != nil {
				log.Errorf("immediate snapshot job failed: %v", err)
				jobs[i].JobFailed = true
				jobs[i].JobError = err.Error()
				break IMMEDIATE_SNAPSHOT
			}

			// snapShottedVM = jobs[i].Snapshot.ResName
			snapshotMap[jobs[i].Snapshot.ResName] = true
			if v.JobType == SchedulerUtils.JOB_TYPE_SNAPSHOT {
				newSnap, removedSnaps, err := zfsutils.TakeScheduledSnapshot(dataset, v.Snapshot.SnapshotType, v.Snapshot.SnapshotsToKeep)
				if err != nil {
					log.Errorf("immediate snapshot job failed: %v", err)
					jobs[i].JobFailed = true
					jobs[i].JobError = err.Error()
				} else {
					log.Infof("new immediate snapshot taken: %s", newSnap)
					log.Infof("old immediate snapshots removed: %v", removedSnaps)
					jobs[i].JobDone = true
				}
			} else if v.JobType == SchedulerUtils.JOB_TYPE_SNAPSHOT_DESTROY {
				err = zfsutils.RemoveSnapshot(jobs[i].Snapshot.SnapshotName)
				if err != nil {
					log.Errorf("snapshot destroy job failed: %v", err)
					jobs[i].JobFailed = true
					jobs[i].JobError = err.Error()
				}
				log.Infof("snapshot destroy job done for: %s", jobs[i].Snapshot.ResName)
				jobs[i].JobDone = true
			} else if v.JobType == SchedulerUtils.JOB_TYPE_SNAPSHOT_ROLLBACK {
				if strings.ToLower(v.ResType) == "vm" {
					// VM
					resOnline, _ := HosterVmUtils.IsVmOnline(jobs[i].Snapshot.ResName)
					if resOnline {
						err = HosterVm.Stop(jobs[i].Snapshot.ResName, true, false)
						if err != nil {
							log.Errorf("snapshot rollback job failed (could not stop the VM): %v", err)
							jobs[i].JobFailed = true
							jobs[i].JobError = err.Error()
							break IMMEDIATE_SNAPSHOT
						}
					}
					// Wait for the VM to stop
					maxTimes := 0
					for resOnline {
						maxTimes++
						if maxTimes > 700 {
							log.Errorf("snapshot rollback job timed-out for %s", jobs[i].Snapshot.ResName)
							jobs[i].JobFailed = true
							jobs[i].JobError = err.Error()
							break IMMEDIATE_SNAPSHOT
						}
						time.Sleep(500 * time.Millisecond)
						resOnline, _ = HosterVmUtils.IsVmOnline(jobs[i].Snapshot.ResName)
					}
				} else {
					// Jail
					err = HosterJail.Stop(jobs[i].Snapshot.ResName)
					if err != nil {
						log.Errorf("snapshot rollback job failed (could not stop the jail): %v", err)
						jobs[i].JobFailed = true
						jobs[i].JobError = err.Error()
						break IMMEDIATE_SNAPSHOT
					}
				}

				// Rollback the snapshot
				err = zfsutils.RollbackSnapshot(jobs[i].Snapshot.SnapshotName)
				if err != nil {
					log.Errorf("snapshot rollback job failed: %v", err)
					jobs[i].JobFailed = true
					jobs[i].JobError = err.Error()
					break IMMEDIATE_SNAPSHOT
				}
				log.Infof("snapshot rollback done for: %s", jobs[i].Snapshot.ResName)
				jobs[i].JobDone = true
			}

			// snapShottedVM = ""
			snapshotMap[jobs[i].Snapshot.ResName] = false
			jobs[i].TimeFinished = time.Now().Unix()
			break IMMEDIATE_SNAPSHOT
		}
	}

	return nil
}
