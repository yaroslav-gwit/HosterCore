package main

import (
	SchedulerUtils "HosterCore/internal/app/scheduler/utils"
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

	replicatedVm string
	// snapShottedVM string
	snapshotMap map[string]bool // this map keeps an exclusive snapshot lock for a specific VM, which prevents snapshot new, snapshot destroy, snapshot replicate and other ZFS conflicts
)

func main() {
	log.Info("starting the scheduler service")
	snapshotMap = make(map[string]bool)
	var wg sync.WaitGroup

	wg.Add(1)
	go socketServer(&wg)

	// We don't care to wait for this routine, because all the jobs will be cleared on exit anyway
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
			executeSnapshotJobs(&jobsMutex)
			time.Sleep(SchedulerUtils.SLEEP_EXECUTE_SNAPSHOTS * time.Second)
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			executeReplicationJobs(&jobsMutex)
			time.Sleep(SchedulerUtils.SLEEP_EXECUTE_REPL * time.Second)
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			executeImmediateSnapshot(&jobsMutex)
			time.Sleep(SchedulerUtils.SLEEP_EXECUTE_IMMEDIATE_SNAPSHOTS * time.Millisecond)
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
			log.Fatal("could not create a UNIX socket:", err)
		}

		go func() {
			err := socketReceive(conn)
			if err != nil {
				log.Errorf("could not handle a new connection %s", err.Error())
			}
		}()
	}
}

var cleanupLogMessage = regexp.MustCompile(`"`)
var cleanupLogMessage2 = regexp.MustCompile(`""`)

func socketReceive(c net.Conn) error {
	defer c.Close()
	log.Infof("new connection [%s]", c.RemoteAddr().Network())

	buffer := make([]byte, 0)
	dynamicBuffer := make([]byte, 1024)

	for {
		bytes, err := c.Read(dynamicBuffer)
		if err != nil {
			log.Errorf("error [%s]", err)
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

	if job.JobType == SchedulerUtils.JOB_TYPE_INFO {
		jobs := getJobs(&jobsMutex)
		// Send the response back through the socket
		resp, err := json.Marshal(jobs)
		if err != nil {
			log.Errorf("could not marshal INFO response: [%s]", err.Error())
		}
		_, err = c.Write(resp)
		if err != nil {
			log.Errorf("could not write the INFO response to socket: [%s]", err.Error())
		} else {
			log.Info("responded with jobs info")
		}
	} else {
		message := strings.TrimSuffix(string(buffer), "\n")
		message = cleanupLogMessage2.ReplaceAllString(message, "nil")
		message = cleanupLogMessage.ReplaceAllString(message, "")

		// Cleanup empty jobs from being logged out
		message = strings.ReplaceAll(message, "replication:{},", "")
		message = strings.ReplaceAll(message, "replication:{}", "")
		message = strings.ReplaceAll(message, "snapshot:{},", "")
		message = strings.ReplaceAll(message, "snapshot:{}", "")
		// EOF Cleanup empty jobs from being logged out

		log.Infof("new job added: [%s]", message)
		addJob(job, &jobsMutex)
	}

	return nil
}

func addJob(job SchedulerUtils.Job, m *sync.RWMutex) error {
	m.Lock()
	defer m.Unlock()

	job.JobId = ulid.Make().String()
	jobs = append(jobs, job)

	return nil
}

// Runs every 25 seconds and removes the old/completed jobs
func removeDoneJobs(m *sync.RWMutex) error {
	m.Lock()
	defer m.Unlock()

	if len(jobs) < 50 {
		return nil
	}

	for i, v := range jobs {
		if v.JobDone && v.JobDoneLogged {
			copy(jobs[i:], jobs[i+1:])
			jobs[len(jobs)-1] = SchedulerUtils.Job{}
			jobs = jobs[0 : len(jobs)-1]

			if v.JobType == SchedulerUtils.JOB_TYPE_REPLICATION {
				log.Infof("replication -> removed the old job for: %s", v.Replication.ResName)
			}
			if v.JobType == SchedulerUtils.JOB_TYPE_SNAPSHOT {
				log.Infof("snapshot -> removed the old job for: %s", v.Snapshot.ResName)
			}

			return nil
		}
	}

	return nil
}

func getJobs(m *sync.RWMutex) (r []SchedulerUtils.Job) {
	m.RLock()
	defer m.RUnlock()
	r = append(r, jobs...)

	return
}

func updateJob(m *sync.RWMutex, job SchedulerUtils.Job) {
	m.Lock()
	defer m.Unlock()

	for i := range jobs {
		if jobs[i].JobId == job.JobId {
			jobs[i] = job
		}
	}
}

// Runs every 6 seconds and executes a first available job
// func executeJobs(m *sync.RWMutex) error {
// 	m.Lock()
// 	defer m.Unlock()

// 	for i, v := range jobs {
// 		if v.JobDone && !v.JobDoneLogged {
// 			if v.JobType == SchedulerUtils.JOB_TYPE_REPLICATION {
// 				logLine := "replication -> done for: " + v.Replication.ResName
// 				log.Info(logLine)
// 			}
// 			if v.JobType == SchedulerUtils.JOB_TYPE_SNAPSHOT {
// 				logLine := "snapshot -> done for: " + v.Snapshot.ResName
// 				log.Info(logLine)
// 			}

// 			jobs[i].JobDoneLogged = true
// 			jobs[i].JobInProgress = false

// 			continue
// 		}

// 		if v.JobFailed && !v.JobFailedLogged {
// 			if v.JobType == SchedulerUtils.JOB_TYPE_REPLICATION {
// 				logLine := "replication -> failed for: " + v.Replication.ResName
// 				log.Error(logLine)
// 			}
// 			if v.JobType == SchedulerUtils.JOB_TYPE_SNAPSHOT && !v.JobFailedLogged {
// 				logLine := "snapshot -> failed for: " + v.Snapshot.ResName
// 				log.Error(logLine)
// 			}

// 			jobs[i].JobFailedLogged = true
// 			jobs[i].JobInProgress = false
// 			continue
// 		}

// 		// If the job is still in progress then break and try again during the next loop
// 		if v.JobInProgress {
// 			if v.JobType == SchedulerUtils.JOB_TYPE_REPLICATION {
// 				jobs[i].JobDone = true
// 				logLine := "replication -> in progress for: " + v.Replication.ResName
// 				log.Info(logLine)
// 			}
// 			if v.JobType == SchedulerUtils.JOB_TYPE_SNAPSHOT {
// 				logLine := "snapshot -> in progress for: " + v.Snapshot.ResName
// 				log.Info(logLine)
// 			}

// 			break
// 		}

// 		if !v.JobDone {
// 			jobs[i].JobInProgress = true

// 			// Replication Jobs
// 			if v.JobType == SchedulerUtils.JOB_TYPE_REPLICATION {
// 				logLine := "replication -> started a new job for: " + v.Replication.ResName
// 				log.Info(logLine)
// 				break
// 			}
// 			// EOF Replication Jobs

// 			// Snapshot Jobs
// 			if v.JobType == SchedulerUtils.JOB_TYPE_SNAPSHOT {
// 				log.Infof("snapshot -> started a new job for: %s", v.Snapshot.ResName)

// 				dataset, err := zfsutils.FindResourceDataset(v.Snapshot.ResName)
// 				if err != nil {
// 					log.Infof("snapshot job jailed: %v", err)
// 					jobs[i].JobFailed = true
// 					jobs[i].JobError = err.Error()
// 				}

// 				newSnap, removedSnaps, err := zfsutils.TakeScheduledSnapshot(dataset, v.Snapshot.SnapshotType, v.Snapshot.SnapshotsToKeep)
// 				if err != nil {
// 					log.Infof("snapshot job jailed: %v", err)
// 					jobs[i].JobFailed = true
// 					jobs[i].JobError = err.Error()
// 				} else {
// 					log.Infof("new snapshot taken: %s", newSnap)
// 					log.Infof("old snapshots removed: %v", removedSnaps)
// 					jobs[i].JobDone = true
// 				}

// 				break
// 			}
// 			// EOF Snapshot Jobs
// 		}
// 	}

// 	return nil
// }
