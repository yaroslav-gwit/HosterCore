package main

import (
	"HosterCore/osfreebsd"
	"encoding/json"
	"log"
	"net"
	"os"
	"strings"
	"sync"
	"time"
)

var SockAddr = "/var/run/hoster_scheduler.sock"

type ReplicationJob struct {
	ZfsDataset       string `json:"zfs_dataset"`
	VmName           string `json:"vm_name"`
	SshEndpoint      string `json:"ssh_endpoint"`
	SshKey           string `json:"ssh_key"`
	BufferSpeedLimit int    `json:"speed_limit"`
	ProgressBytes    int    `json:"progress_bytes"`
	ProgressPercent  int    `json:"progress_percent"`
}

type SnapshotJob struct {
	ZfsDataset      string `json:"zfs_dataset"`
	VmName          string `json:"vm_name"`
	SnapshotsToKeep int    `json:"snapshots_to_keep"`
	SnapshotType    string `json:"snapshot_type"`
}

const (
	JOB_TYPE_REPLICATION = "replication"
	JOB_TYPE_SNAPSHOT    = "snapshot"

	SLEEP_REMOVE_DONE_JOBS = 17 // used as seconds in the removeDoneJobs loop
	SLEEP_EXECUTE_JOBS     = 6  // used as seconds in the executeJobs loop
)

type Job struct {
	JobDone         bool           `json:"job_done"`
	JobDoneLogged   bool           `json:"job_done_logged"`
	JobNext         bool           `json:"job_next"`
	JobInProgress   bool           `json:"job_in_progress"`
	JobFailed       bool           `json:"job_failed"`
	JobFailedLogged bool           `json:"job_failed_logged"`
	JobError        string         `json:"job_error"`
	JobType         string         `json:"job_type"`
	Replication     ReplicationJob `json:"replication"`
	Snapshot        SnapshotJob    `json:"snapshot"`
}

var (
	jobs      = []Job{}
	jobsMutex sync.RWMutex
)

func main() {
	var wg sync.WaitGroup

	wg.Add(1)
	go socketServer(&wg)

	go func() {
		for {
			removeDoneJobs(&jobsMutex)
			time.Sleep(SLEEP_REMOVE_DONE_JOBS * time.Second)
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			executeJobs(&jobsMutex)
			time.Sleep(SLEEP_EXECUTE_JOBS * time.Second)
		}
	}()

	wg.Wait()
}

func socketServer(wg *sync.WaitGroup) {
	if err := os.RemoveAll(SockAddr); err != nil {
		log.Fatal(err)
	}

	newSocket, err := net.Listen("unix", SockAddr)
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
				log.Println(err)
			}
		}()
	}
}

func socketReceive(c net.Conn) error {
	log.Printf("Client connected [%s]", c.RemoteAddr().Network())

	buffer := make([]byte, 0)
	dynamicBuffer := make([]byte, 1024)

	for {
		bytes, err := c.Read(dynamicBuffer)
		if err != nil {
			log.Printf("Error [%s]", err)
			break
		}

		buffer = append(buffer, dynamicBuffer[0:bytes]...)

		if bytes < len(dynamicBuffer) {
			break
		}
	}

	message := strings.TrimSuffix(string(buffer), "\n")

	job := Job{}
	err := json.Unmarshal(buffer, &job)
	if err != nil {
		return err
	}
	addJob(job, &jobsMutex)

	log.Printf("Client has sent a message [%s]", message)
	defer c.Close()
	return nil
}

func addJob(job Job, m *sync.RWMutex) error {
	m.Lock()
	defer m.Unlock()

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
			jobs[len(jobs)-1] = Job{}
			jobs = jobs[0 : len(jobs)-1]

			if v.JobType == JOB_TYPE_REPLICATION {
				logLine := "Replication -> Removed the old job for: " + v.Replication.VmName
				go osfreebsd.LoggerToSyslog(osfreebsd.LOGGER_SRV_SCHEDULER, osfreebsd.LOGGER_LEVEL_DEBUG, logLine)
				log.Println(logLine)
			}
			if v.JobType == JOB_TYPE_SNAPSHOT {
				logLine := "Snapshot -> Removed the old job for: " + v.Snapshot.VmName
				go osfreebsd.LoggerToSyslog(osfreebsd.LOGGER_SRV_SCHEDULER, osfreebsd.LOGGER_LEVEL_DEBUG, logLine)
				log.Println(logLine)
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
			if v.JobType == JOB_TYPE_REPLICATION {
				logLine := "Replication -> Done for: " + v.Replication.VmName
				go osfreebsd.LoggerToSyslog(osfreebsd.LOGGER_SRV_SCHEDULER, osfreebsd.LOGGER_LEVEL_INFO, logLine)
				log.Println(logLine)
			}
			if v.JobType == JOB_TYPE_SNAPSHOT {
				logLine := "Snapshot -> Done for: " + v.Snapshot.VmName
				go osfreebsd.LoggerToSyslog(osfreebsd.LOGGER_SRV_SCHEDULER, osfreebsd.LOGGER_LEVEL_INFO, logLine)
				log.Println(logLine)
			}

			jobs[i].JobDoneLogged = true
			jobs[i].JobInProgress = false

			continue
		}

		if v.JobFailed && !v.JobFailedLogged {
			if v.JobType == JOB_TYPE_REPLICATION {
				logLine := "Replication -> Failed for: " + v.Replication.VmName
				go osfreebsd.LoggerToSyslog(osfreebsd.LOGGER_SRV_SCHEDULER, osfreebsd.LOGGER_LEVEL_INFO, logLine)
				log.Println(logLine)
			}
			if v.JobType == JOB_TYPE_SNAPSHOT && !v.JobFailedLogged {
				logLine := "Snapshot -> Failed for: " + v.Snapshot.VmName
				go osfreebsd.LoggerToSyslog(osfreebsd.LOGGER_SRV_SCHEDULER, osfreebsd.LOGGER_LEVEL_INFO, logLine)
				log.Println(logLine)
			}

			jobs[i].JobFailedLogged = true
			jobs[i].JobInProgress = false
			continue
		}

		// If the job is still in progress then break and try again during the next loop
		if v.JobInProgress {
			if v.JobType == JOB_TYPE_REPLICATION {
				jobs[i].JobDone = true
				logLine := "Replication -> In progress for: " + v.Replication.VmName
				go osfreebsd.LoggerToSyslog(osfreebsd.LOGGER_SRV_SCHEDULER, osfreebsd.LOGGER_LEVEL_INFO, logLine)
				log.Println(logLine)
			}
			if v.JobType == JOB_TYPE_SNAPSHOT {
				jobs[i].JobDone = true
				logLine := "Snapshot -> In progress for: " + v.Snapshot.VmName
				go osfreebsd.LoggerToSyslog(osfreebsd.LOGGER_SRV_SCHEDULER, osfreebsd.LOGGER_LEVEL_INFO, logLine)
				log.Println(logLine)
			}

			break
		}

		if !v.JobDone {
			jobs[i].JobInProgress = true

			if v.JobType == JOB_TYPE_REPLICATION {
				// replicate
				logLine := "Replication -> Added a new job for: " + v.Replication.VmName
				go osfreebsd.LoggerToSyslog(osfreebsd.LOGGER_SRV_SCHEDULER, osfreebsd.LOGGER_LEVEL_INFO, logLine)
				log.Println(logLine)
				break
			}

			if v.JobType == JOB_TYPE_SNAPSHOT {
				// snapshot
				logLine := "Snapshot -> Added a new job for: " + v.Snapshot.VmName
				go osfreebsd.LoggerToSyslog(osfreebsd.LOGGER_SRV_SCHEDULER, osfreebsd.LOGGER_LEVEL_INFO, logLine)
				log.Println(logLine)
				break
			}
		}
	}

	return nil
}
