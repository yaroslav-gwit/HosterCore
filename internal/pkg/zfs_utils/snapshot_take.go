// Copyright 2023 Hoster Authors. All rights reserved.
// Use of this source code is governed by an Apache License 2.0
// license that can be found in the LICENSE file.

package zfsutils

import (
	"fmt"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"golang.org/x/exp/slices"
)

// Takes a new snapshot, returns the name of the new snapshot, list of the removed snapshots, and/or an error.
//
// Useful for scheduling the automated snapshot jobs.
func TakeScheduledSnapshot(dataset string, snapshotType string, keep int) (snapshotName string, removedSnapshots []string, e error) {
	snapshotTypes := []string{"replication", "custom", "frequent", "hourly", "daily", "weekly", "monthly", "yearly"}
	if slices.Contains(snapshotTypes, snapshotType) {
		_ = 0
	} else {
		e = fmt.Errorf("please provide the correct snapshot type")
		return
	}

	timeNow := time.Now().Format("20060102_150405.000000")
	snapshotName = dataset + "@" + snapshotType + "_" + timeNow

	out, err := exec.Command("zfs", "snapshot", snapshotName).CombinedOutput()
	if err != nil {
		e = fmt.Errorf(strings.TrimSpace(string(out)) + "; " + err.Error())
		return
	}

	reSnapTypeMatch := regexp.MustCompile(`@` + snapshotType + "_")
	datasetSnapshots := []SnapshotInfo{}
	// removedSnapshots := []string{}
	allSnapshots, err := SnapshotListAll()
	if err != nil {
		e = err
		return
	}

	for _, v := range allSnapshots {
		if v.Dataset == dataset {
			if v.Locked || len(v.Clones) > 0 {
				continue
			}
			if reSnapTypeMatch.MatchString(v.Name) {
				datasetSnapshots = append(datasetSnapshots, v)
			}
		}
	}

	if len(datasetSnapshots) <= keep {
		return
	}

	for i := 0; i < len(datasetSnapshots)-keep; i++ {
		err := RemoveSnapshot(datasetSnapshots[i].Name)
		if err != nil {
			e = err
			return
		}
		removedSnapshots = append(removedSnapshots, datasetSnapshots[i].Name)
	}

	return snapshotName, removedSnapshots, nil
}
