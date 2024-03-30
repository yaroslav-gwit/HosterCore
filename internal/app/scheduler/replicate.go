package main

import (
	SchedulerUtils "HosterCore/internal/app/scheduler/utils"
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"github.com/oklog/ulid/v2"
)

// Runs every 5 seconds and executes a first available job
func executeReplicationJobs(m *sync.RWMutex) error {
	m.Lock()
	defer m.Unlock()

	if len(replicatedVm) > 0 {
		return nil
	}

	for i, v := range jobs {
		if v.JobType != SchedulerUtils.JOB_TYPE_REPLICATION {
			continue
		}

		if v.JobDone && !v.JobDoneLogged {
			logLine := "replication -> done for: " + v.Replication.ResName
			log.Info(logLine)
			jobs[i].JobDoneLogged = true
			jobs[i].JobInProgress = false
			continue
		}

		if v.JobFailed && !v.JobFailedLogged {
			logLine := "replication -> failed for: " + v.Replication.ResName
			log.Error(logLine)
			jobs[i].JobFailedLogged = true
			jobs[i].JobInProgress = false
			continue
		}

		// If the job is still in progress then break and try again during the next loop
		if v.JobInProgress {
			jobs[i].JobDone = true
			logLine := "replication -> in progress for: " + v.Replication.ResName
			log.Info(logLine)

			break
		}

		if !v.JobDone {
			jobs[i].JobInProgress = true

			if v.JobType == SchedulerUtils.JOB_TYPE_REPLICATION {
				logLine := "replication -> started a new job for: " + v.Replication.ResName
				log.Info(logLine)
				break
			}
		}
	}

	return nil
}

func Replicate(job SchedulerUtils.Job, wg *sync.WaitGroup) error {
	replicatedVm = job.Replication.ResName
	scriptsToRemove := []string{}
	defer func() {
		replicatedVm = ""
		job.JobDone = true
		for _, v := range scriptsToRemove {
			os.Remove(v)
		}
	}()

	reMatchSize := regexp.MustCompile(`^size.*`)
	reMatchSpace := regexp.MustCompile(`\s+`)
	reMatchTime := regexp.MustCompile(`.*\d\d:\d\d:\d\d.*`)

	for _, v := range job.Replication.ScriptsRemove {
		destroyFile := "/tmp/" + ulid.Make().String()
		err := os.WriteFile(destroyFile, []byte(v), 0600)
		if err != nil {
			job.JobError = err.Error()
			updateJob(&jobsMutex, job)
			return err
		}
		scriptsToRemove = append(scriptsToRemove, destroyFile)

		out, err := exec.Command("sh", destroyFile).CombinedOutput()
		if err != nil {
			job.JobError = err.Error()
			updateJob(&jobsMutex, job)
			return fmt.Errorf("%s; %s", strings.TrimSpace(string(out)), err.Error())
		}
	}

	for _, v := range job.Replication.ScriptsReplicate {
		replFile := "/tmp/" + ulid.Make().String()
		err := os.WriteFile(replFile, []byte(v), 0600)
		if err != nil {
			job.JobError = err.Error()
			updateJob(&jobsMutex, job)
			return err
		}
		scriptsToRemove = append(scriptsToRemove, replFile)

		cmd := exec.Command("sh", replFile)
		stderr, err := cmd.StderrPipe()
		if err != nil {
			job.JobError = err.Error()
			updateJob(&jobsMutex, job)
			return err
		}

		if err := cmd.Start(); err != nil {
			job.JobError = err.Error()
			updateJob(&jobsMutex, job)
			return err
		}

		scanner := bufio.NewScanner(stderr)
		errLines := []string{}
		for scanner.Scan() {
			line := scanner.Text()
			if reMatchSize.MatchString(line) {
				temp, err := strconv.ParseUint(reMatchSpace.Split(line, -1)[1], 10, 64)
				if err != nil {
					job.JobError = err.Error()
					updateJob(&jobsMutex, job)
					return err
				}
				// emojlog.PrintLogMessage("Snapshot size: "+byteconversion.BytesToHuman(temp), emojlog.Debug)
				job.Replication.ProgressBytesTotal = temp
				updateJob(&jobsMutex, job)
			} else if reMatchTime.MatchString(line) {
				temp, err := strconv.ParseUint(reMatchSpace.Split(line, -1)[1], 10, 64)
				if err != nil {
					job.JobError = err.Error()
					updateJob(&jobsMutex, job)
					return err
				}
				job.Replication.ProgressBytesDone = temp
				updateJob(&jobsMutex, job)
				// fmt.Printf("Copied so far: %d\n", temp)
			} else {
				errLines = append(errLines, line)
			}
		}

		// Wait for command to finish
		err = cmd.Wait()
		if err != nil {
			return fmt.Errorf("%v", errLines)
		}
	}

	return nil
}
