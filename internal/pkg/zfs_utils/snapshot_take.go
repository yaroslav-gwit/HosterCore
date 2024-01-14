package zfsutils

import (
	"errors"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"golang.org/x/exp/slices"
)

// Takes a new snapshot, and returns the name of the new snapshot, list of the removed snapshots, or an error
// Useful for scheduling the automated snapshot jobs
func TakeScheduledSnapshot(dataset string, snapshotType string, keep int) (string, []string, error) {
	snapshotTypes := []string{"replication", "custom", "frequent", "hourly", "daily", "weekly", "monthly", "yearly"}
	if slices.Contains(snapshotTypes, snapshotType) {
		_ = 0
	} else {
		return "", []string{}, errors.New("please provide the correct snapshot type")
	}

	timeNow := time.Now().Format("2006-01-02_15-04-05.000")
	snapshotName := dataset + "@" + snapshotType + "_" + timeNow

	out, err := exec.Command("zfs", "snapshot", snapshotName).CombinedOutput()
	if err != nil {
		errString := strings.TrimSpace(string(out)) + "; " + err.Error()
		return "", []string{}, errors.New(errString)
	}

	reSnapTypeMatch := regexp.MustCompile(`@` + snapshotType + "_")

	datasetSnapshots := []SnapshotInfo{}
	removedSnapshots := []string{}
	allSnapshots, err := SnapshotListAll()
	if err != nil {
		return "", []string{}, err
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
		return snapshotName, []string{}, nil
	}

	for i := 0; i < len(datasetSnapshots)-keep; i++ {
		err := RemoveSnapshot(datasetSnapshots[i].Name)
		if err != nil {
			return "", []string{}, err
		}
		removedSnapshots = append(removedSnapshots, datasetSnapshots[i].Name)
	}

	return snapshotName, removedSnapshots, nil
}
