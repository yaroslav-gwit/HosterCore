// Copyright 2024 Hoster Authors. All rights reserved.
// Use of this source code is governed by an Apache License 2.0
// license that can be found in the LICENSE file.

package SchedulerClient

import (
	SchedulerUtils "HosterCore/internal/app/scheduler/utils"
	"bufio"
	"encoding/json"
	"fmt"
	"net"
)

func GetJobInfo(jobID string) (r SchedulerUtils.Job, e error) {
	c, err := net.Dial("unix", SchedulerUtils.SockAddr)
	if err != nil {
		e = err
		return
	}
	defer c.Close()

	var job SchedulerUtils.Job
	job.JobType = SchedulerUtils.JOB_TYPE_INFO

	jsonJob, err := json.Marshal(job)
	if err != nil {
		e = err
		return
	}

	jsonJob = append(jsonJob, '\n')
	_, err = c.Write(jsonJob)
	if err != nil {
		e = err
		return
	}

	// Read the response from the socket
	reader := bufio.NewReader(c)
	jsonResponse, err := reader.ReadBytes('\n')
	if err != nil {
		e = err
		return
	}
	jsonResponse = jsonResponse[:len(jsonResponse)-1]

	// Process the JSON response as needed
	var jobs []SchedulerUtils.Job
	err = json.Unmarshal(jsonResponse, &jobs)
	if err != nil {
		e = err
		return
	}

	for i := range jobs {
		if jobs[i].JobId == jobID {
			r = jobs[i]
			return
		}
	}

	e = fmt.Errorf("could not find the job using the ID provided")
	return
}

func GetJobList() (r []SchedulerUtils.Job, e error) {
	c, err := net.Dial("unix", SchedulerUtils.SockAddr)
	if err != nil {
		e = err
		return
	}
	defer c.Close()

	var job SchedulerUtils.Job
	job.JobType = SchedulerUtils.JOB_TYPE_INFO

	jsonJob, err := json.Marshal(job)
	if err != nil {
		e = err
		return
	}

	jsonJob = append(jsonJob, '\n')
	_, err = c.Write(jsonJob)
	if err != nil {
		e = err
		return
	}

	// Read the response from the socket
	reader := bufio.NewReader(c)
	jsonResponse, err := reader.ReadBytes('\n')
	if err != nil {
		e = err
		return
	}
	jsonResponse = jsonResponse[:len(jsonResponse)-1]

	// Process the JSON response as needed
	err = json.Unmarshal(jsonResponse, &r)
	if err != nil {
		e = err
		return
	}

	return
}
