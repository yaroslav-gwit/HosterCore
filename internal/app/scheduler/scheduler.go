package main

import (
	SchedulerUtils "HosterCore/internal/app/scheduler/utils"
	zfsutils "HosterCore/internal/pkg/zfs_utils"
	"encoding/json"
	"net"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/oklog/ulid/v2"
)

var (
	jobs      = []SchedulerUtils.Job{}
	jobsMutex sync.RWMutex
)

func main() {
	log.Info("Starting the Scheduler service")
	var wg sync.WaitGroup

	wg.Add(1)
	go socketServer(&wg)

	go func() {
		for {
			removeDoneJobs(&jobsMutex)
			time.Sleep(SchedulerUtils.SLEEP_REMOVE_DONE_JOBS * time.Second)
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			executeJobs(&jobsMutex)
			time.Sleep(SchedulerUtils.SLEEP_EXECUTE_JOBS * time.Second)
		}
	}()

	wg.Wait()
}

func socketServer(wg *sync.WaitGroup) {
	if err := os.RemoveAll(SchedulerUtils.SockAddr); err != nil {
		log.Fatal(err)
	}

	newSocket, err := net.Listen("unix", SchedulerUtils.SockAddr)
	if err != nil {
		log.Fatal("listen error:", err)
	}

	defer wg.Done()
	defer newSocket.Close()

	for {
		conn, err := newSocket.Accept()
		if err != nil {
			log.Fatal("accept error:", err)
		}

		go func() {
			err := socketReceive(conn)
			if err != nil {
				log.Error(err)
			}
		}()
	}
}

var cleanupLogMessage = regexp.MustCompile(`"`)
var cleanupLogMessage2 = regexp.MustCompile(`""`)

func socketReceive(c net.Conn) error {
	log.Infof("Client connected [%s]", c.RemoteAddr().Network())

	buffer := make([]byte, 0)
	dynamicBuffer := make([]byte, 1024)

	for {
		bytes, err := c.Read(dynamicBuffer)
		if err != nil {
			log.Errorf("Error [%s]", err)
			break
		}

		buffer = append(buffer, dynamicBuffer[0:bytes]...)

		if bytes < len(dynamicBuffer) {
			break
		}
	}

	job := SchedulerUtils.Job{}
	err := json.Unmarshal(buffer, &job)
	if err != nil {
		return err
	}
	addJob(job, &jobsMutex)

	message := strings.TrimSuffix(string(buffer), "\n")
	message = cleanupLogMessage2.ReplaceAllString(message, "nil")
	message = cleanupLogMessage.ReplaceAllString(message, "")
	log.Infof("Client has sent a message: [%s]", message)
	defer c.Close()
	return nil
}

func addJob(job SchedulerUtils.Job, m *sync.RWMutex) error {
	m.Lock()
	defer m.Unlock()

	job.JobId = ulid.Make().String()
	jobs = append(jobs, job)

	return nil
}

// func getJobs(m *sync.RWMutex) ([]Job, error) {
// 	m.RLock()
// 	defer m.RUnlock()

// 	jobsCopy := []Job{}
// 	_ = copy(jobsCopy, jobs)

// 	return jobsCopy, nil
// }

// Runs every 17 seconds and removes the old/completed jobs
func removeDoneJobs(m *sync.RWMutex) error {
	m.Lock()
	defer m.Unlock()

	if len(jobs) < 30 {
		return nil
	}

	for i, v := range jobs {
		if v.JobDone && v.JobDoneLogged {
			copy(jobs[i:], jobs[i+1:])
			jobs[len(jobs)-1] = SchedulerUtils.Job{}
			jobs = jobs[0 : len(jobs)-1]

			if v.JobType == SchedulerUtils.JOB_TYPE_REPLICATION {
				logLine := "Replication -> Removed the old job for: " + v.Replication.VmName
				// go osfreebsd.LoggerToSyslog(osfreebsd.LOGGER_SRV_SCHEDULER, osfreebsd.LOGGER_LEVEL_DEBUG, logLine)
				log.Info(logLine)
			}
			if v.JobType == SchedulerUtils.JOB_TYPE_SNAPSHOT {
				logLine := "Snapshot -> Removed the old job for: " + v.Snapshot.VmName
				// go osfreebsd.LoggerToSyslog(osfreebsd.LOGGER_SRV_SCHEDULER, osfreebsd.LOGGER_LEVEL_DEBUG, logLine)
				log.Info(logLine)
			}

			return nil
		}
	}

	return nil
}

// Runs every 6 seconds and executes a first available job
func executeJobs(m *sync.RWMutex) error {
	m.Lock()
	defer m.Unlock()

	for i, v := range jobs {
		if v.JobDone && !v.JobDoneLogged {
			if v.JobType == SchedulerUtils.JOB_TYPE_REPLICATION {
				logLine := "Replication -> Done for: " + v.Replication.VmName
				// go osfreebsd.LoggerToSyslog(osfreebsd.LOGGER_SRV_SCHEDULER, osfreebsd.LOGGER_LEVEL_INFO, logLine)
				log.Info(logLine)
			}
			if v.JobType == SchedulerUtils.JOB_TYPE_SNAPSHOT {
				logLine := "Snapshot -> Done for: " + v.Snapshot.VmName
				// go osfreebsd.LoggerToSyslog(osfreebsd.LOGGER_SRV_SCHEDULER, osfreebsd.LOGGER_LEVEL_INFO, logLine)
				log.Info(logLine)
			}

			jobs[i].JobDoneLogged = true
			jobs[i].JobInProgress = false

			continue
		}

		if v.JobFailed && !v.JobFailedLogged {
			if v.JobType == SchedulerUtils.JOB_TYPE_REPLICATION {
				logLine := "Replication -> Failed for: " + v.Replication.VmName
				// go osfreebsd.LoggerToSyslog(osfreebsd.LOGGER_SRV_SCHEDULER, osfreebsd.LOGGER_LEVEL_INFO, logLine)
				log.Error(logLine)
			}
			if v.JobType == SchedulerUtils.JOB_TYPE_SNAPSHOT && !v.JobFailedLogged {
				logLine := "Snapshot -> Failed for: " + v.Snapshot.VmName
				// go osfreebsd.LoggerToSyslog(osfreebsd.LOGGER_SRV_SCHEDULER, osfreebsd.LOGGER_LEVEL_INFO, logLine)
				log.Error(logLine)
			}

			jobs[i].JobFailedLogged = true
			jobs[i].JobInProgress = false
			continue
		}

		// If the job is still in progress then break and try again during the next loop
		if v.JobInProgress {
			if v.JobType == SchedulerUtils.JOB_TYPE_REPLICATION {
				jobs[i].JobDone = true
				logLine := "Replication -> In progress for: " + v.Replication.VmName
				// go osfreebsd.LoggerToSyslog(osfreebsd.LOGGER_SRV_SCHEDULER, osfreebsd.LOGGER_LEVEL_INFO, logLine)
				log.Info(logLine)
			}
			if v.JobType == SchedulerUtils.JOB_TYPE_SNAPSHOT {
				logLine := "Snapshot -> In progress for: " + v.Snapshot.VmName
				// go osfreebsd.LoggerToSyslog(osfreebsd.LOGGER_SRV_SCHEDULER, osfreebsd.LOGGER_LEVEL_INFO, logLine)
				log.Info(logLine)
			}

			break
		}

		if !v.JobDone {
			jobs[i].JobInProgress = true

			if v.JobType == SchedulerUtils.JOB_TYPE_REPLICATION {
				// replicate
				logLine := "Replication -> Started a new job for: " + v.Replication.VmName
				// go osfreebsd.LoggerToSyslog(osfreebsd.LOGGER_SRV_SCHEDULER, osfreebsd.LOGGER_LEVEL_INFO, logLine)
				log.Info(logLine)
				break
			}

			if v.JobType == SchedulerUtils.JOB_TYPE_SNAPSHOT {
				// snapshot
				logLine := "Snapshot -> Started a new job for: " + v.Snapshot.VmName
				// go osfreebsd.LoggerToSyslog(osfreebsd.LOGGER_SRV_SCHEDULER, osfreebsd.LOGGER_LEVEL_INFO, logLine)
				log.Info(logLine)

				dataset, err := zfsutils.FindResourceDataset(v.Snapshot.VmName)
				if err != nil {
					log.Infof("Snapshot job jailed: %v", err)
					jobs[i].JobFailed = true
					jobs[i].JobError = err.Error()
				}

				newSnap, removedSnaps, err := zfsutils.TakeScheduledSnapshot(dataset, v.Snapshot.SnapshotType, v.Snapshot.SnapshotsToKeep)
				if err != nil {
					log.Infof("Snapshot job jailed: %v", err)
					jobs[i].JobFailed = true
					jobs[i].JobError = err.Error()
				} else {
					log.Infof("New snapshot taken: %s", newSnap)
					log.Infof("Removed snapshots: %v", removedSnaps)
					jobs[i].JobDone = true
				}

				break
			}
		}
	}

	return nil
}
