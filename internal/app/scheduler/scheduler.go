package main

import (
	SchedulerUtils "HosterCore/internal/app/scheduler/utils"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/oklog/ulid/v2"
)

var (
	jobs              = []SchedulerUtils.Job{}
	jobsMutex         = &sync.RWMutex{}
	snapshotMap       map[string]bool // this map keeps an exclusive snapshot lock for a specific VM, which prevents snapshot new, snapshot destroy, snapshot replicate and other ZFS conflicts
	replicatedVm      string
	replicatedVmMutex = &sync.RWMutex{}
)

var version = "" // automatically set during the build process

func main() {
	// Print the version and exit
	args := os.Args
	if len(args) > 1 {
		res := os.Args[1]
		if res == "version" || res == "v" || res == "--version" || res == "-v" {
			fmt.Println(version)
			return
		}
	}

	log.Info("starting the scheduler service")
	snapshotMap = make(map[string]bool)
	var wg sync.WaitGroup

	wg.Add(1)
	go socketServer(&wg)

	// We don't care to wait for this routine, because all the jobs will be cleared on exit anyway
	go func() {
		for {
			removeDoneJobs(jobsMutex)
			time.Sleep(SchedulerUtils.SLEEP_REMOVE_DONE_JOBS * time.Second)
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			executeSnapshotJobs(jobsMutex)
			time.Sleep(SchedulerUtils.SLEEP_EXECUTE_SNAPSHOTS * time.Second)
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			executeReplicationJobs(jobsMutex)
			time.Sleep(SchedulerUtils.SLEEP_EXECUTE_REPL * time.Second)
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			executeImmediateSnapshot(jobsMutex)
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
		jobs := getJobs(jobsMutex)
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
		addJob(job, jobsMutex)
	}

	return nil
}

func addJob(job SchedulerUtils.Job, m *sync.RWMutex) error {
	m.Lock()
	defer m.Unlock()

	if len(job.JobId) < 1 {
		job.JobId = ulid.Make().String()
	}
	job.TimeAdded = time.Now().Unix()

	// Only add the replication job if the resource is not already being replicated
	if job.JobType == SchedulerUtils.JOB_TYPE_REPLICATION {
		for _, v := range jobs {
			if v.Replication.ResName == job.Replication.ResName && !v.JobDone {
				log.Warnf("resource %s is already being replicated, new replication job will be ignored", job.Replication.ResName)
				return nil
			}
		}
	}

	// Only add the snapshot job if the resource is not already being replicated
	if job.JobType == SchedulerUtils.JOB_TYPE_SNAPSHOT {
		for _, v := range jobs {
			if v.Replication.ResName == job.Snapshot.ResName && !v.JobDone {
				log.Warnf("resource %s is being replicated, new snapshot job will be ignored", job.Snapshot.ResName)
				return nil
			}
		}
	}

	// Append the job to the jobs slice
	jobs = append(jobs, job)
	return nil
}

// Runs every 10 seconds and removes the old/completed jobs
func removeDoneJobs(m *sync.RWMutex) error {
	m.Lock()
	defer m.Unlock()

	if len(jobs) < 50 {
		return nil
	}

	for i, v := range jobs {
		if v.JobDone && v.JobDoneLogged {
			// Remove a single Job
			copy(jobs[i:], jobs[i+1:])
			jobs[len(jobs)-1] = SchedulerUtils.Job{}
			jobs = jobs[0 : len(jobs)-1]

			log.Infof("removed an old job for: %s, with job id: %s", v.Snapshot.ResName, v.JobId)
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
