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

				go func(input SchedulerUtils.Job) {
					err := Replicate(input, m)
					if err != nil {
						input.JobFailed = true
						input.JobError = err.Error()
						updateJob(m, input)
					}
				}(jobs[i])
			}
		}
	}

	return nil
}

func Replicate(job SchedulerUtils.Job, m *sync.RWMutex) error {
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
			return err
		}
		scriptsToRemove = append(scriptsToRemove, destroyFile)

		out, err := exec.Command("sh", destroyFile).CombinedOutput()
		if err != nil {
			return fmt.Errorf("%s; %s", strings.TrimSpace(string(out)), err.Error())
		}
	}

	for i, v := range job.Replication.ScriptsReplicate {
		replFile := "/tmp/" + ulid.Make().String()
		err := os.WriteFile(replFile, []byte(v), 0600)
		if err != nil {
			return err
		}
		scriptsToRemove = append(scriptsToRemove, replFile)

		cmd := exec.Command("sh", replFile)
		stderr, err := cmd.StderrPipe()
		if err != nil {
			return err
		}

		if err := cmd.Start(); err != nil {
			return err
		}

		job.Replication.ProgressTotalSnaps = len(job.Replication.ScriptsReplicate)
		updateJob(m, job)

		scanner := bufio.NewScanner(stderr)
		errLines := []string{}
		for scanner.Scan() {
			line := scanner.Text()
			if reMatchSize.MatchString(line) {
				temp, err := strconv.ParseUint(reMatchSpace.Split(line, -1)[1], 10, 64)
				if err != nil {
					return err
				}
				// emojlog.PrintLogMessage("Snapshot size: "+byteconversion.BytesToHuman(temp), emojlog.Debug)
				job.Replication.ProgressBytesTotal = temp
				updateJob(m, job)
			} else if reMatchTime.MatchString(line) {
				temp, err := strconv.ParseUint(reMatchSpace.Split(line, -1)[1], 10, 64)
				if err != nil {
					return err
				}
				job.Replication.ProgressBytesDone = temp
				updateJob(m, job)
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

		job.Replication.ProgressDoneSnaps = i + 1
		updateJob(m, job)
	}

	return nil
}
