// Copyright 2024 Hoster Authors. All rights reserved.
// Use of this source code is governed by an Apache License 2.0
// license that can be found in the LICENSE file.

package SchedulerUtils

const SockAddr = "/var/run/hoster_scheduler.sock"

const JOB_TYPE_REPLICATION = "replication"
const JOB_TYPE_SNAPSHOT = "snapshot"
const JOB_TYPE_INFO = "info"

const SLEEP_REMOVE_DONE_JOBS = 9 // used as seconds in the removeDoneJobs loop
const SLEEP_EXECUTE_JOBS = 4     // used as seconds in the executeJobs loop
