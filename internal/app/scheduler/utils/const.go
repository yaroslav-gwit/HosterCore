// Copyright 2024 Hoster Authors. All rights reserved.
// Use of this source code is governed by an Apache License 2.0
// license that can be found in the LICENSE file.

package SchedulerUtils

const SockAddr = "/var/run/hoster_scheduler.sock"

const JOB_TYPE_SNAPSHOT_ROLLBACK = "snapshot_rollback"
const JOB_TYPE_SNAPSHOT_DESTROY = "snapshot_destroy"
const JOB_TYPE_REPLICATION = "replication"
const JOB_TYPE_SNAPSHOT = "snapshot"
const JOB_TYPE_INFO = "info"

const SLEEP_REMOVE_DONE_JOBS = 10             // used as seconds in the removeDoneJobs loop
const SLEEP_EXECUTE_SNAPSHOTS = 5             // used as seconds in the executeSnapshotJobs loop
const SLEEP_EXECUTE_IMMEDIATE_SNAPSHOTS = 500 // used as milliseconds in the executeImmediateSnapshotJobs loop
const SLEEP_EXECUTE_REPL = 5                  // used as seconds in the executeReplicationJobs loop
