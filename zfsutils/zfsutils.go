package zfsutils

import (
	"HosterCore/osfreebsd"
	"errors"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"

	"golang.org/x/exp/slices"
)

type SnapshotInfo struct {
	Name        string   `json:"snapshot_name"`
	Dataset     string   `json:"snapshot_dataset"`
	ShortName   string   `json:"snapshot_short_name"`
	Locked      bool     `json:"snapshot_locked"`
	Clones      []string `json:"snapshot_clones"`
	SizeBytes   uint64   `json:"snapshot_size_bytes"`
	SizeHuman   string   `json:"snapshot_size_human"`
	Description string   `json:"snapshot_description"`
}

// Returns all ZFS snapshots present on this system
func SnapshotListAll() ([]SnapshotInfo, error) {
	info := []SnapshotInfo{}

	reSplitSpace := regexp.MustCompile(`\s+`)
	out, err := exec.Command("/sbin/zfs", "list", "-t", "snapshot", "-p", "-o", "name,used,userrefs,clones").CombinedOutput()
	if err != nil {
		errString := strings.TrimSpace(string(out)) + "; " + err.Error()
		return []SnapshotInfo{}, errors.New(errString)
	}

	// Example output
	// NAME                                                                                  USED  USERREFS  CLONES
	// zroot/vm-encrypted/test-vm-0107@hourly_2023-11-29_19-33-00                         2584576  0         zroot/vm-encrypted/cloneMe2,zroot/vm-encrypted/cloneMe1
	// zroot/vm-encrypted/test-vm-0106@custom_2023-08-14_15-53-25                               0  0         -
	nameIndex := -1
	usedIndex := -1
	userRefsIndex := -1
	clonesIndex := -1

	// Parse the header
	for i, v := range strings.Split(string(out), "\n") {
		if i == 0 {
			for ii, vv := range reSplitSpace.Split(v, -1) {
				if strings.TrimSpace(vv) == "NAME" {
					nameIndex = ii
				} else if strings.TrimSpace(vv) == "USED" {
					usedIndex = ii
				} else if strings.TrimSpace(vv) == "USERREFS" {
					userRefsIndex = ii
				} else if strings.TrimSpace(vv) == "CLONES" {
					clonesIndex = ii
				}
			}
		}
	}

	// Check if the header was parsed correctly
	if nameIndex == -1 {
		return []SnapshotInfo{}, errors.New("could not parse a name index")
	}
	if usedIndex == -1 {
		return []SnapshotInfo{}, errors.New("could not parse a used index")
	}
	if userRefsIndex == -1 {
		return []SnapshotInfo{}, errors.New("could not parse a user refs index")
	}
	if clonesIndex == -1 {
		return []SnapshotInfo{}, errors.New("could not parse a cloned index")
	}

	// Parse the output without a header
	for i, v := range strings.Split(string(out), "\n") {
		if i == 0 {
			continue
		}
		if len(v) < 1 {
			continue
		}

		infoTemp := SnapshotInfo{}
		tmpList := reSplitSpace.Split(v, -1)

		if len(tmpList) < 1 {
			continue
		}

		nameSplit := strings.Split(tmpList[nameIndex], "@")
		if len(nameSplit) < 2 {
			continue
		} else {
			infoTemp.Dataset = nameSplit[0]
			infoTemp.Name = tmpList[nameIndex]
			infoTemp.ShortName = nameSplit[1]
		}

		if tmpList[userRefsIndex] == "0" {
			infoTemp.Locked = false
		} else {
			infoTemp.Locked = true
		}

		if tmpList[clonesIndex] == "-" {
			_ = 0
		} else {
			infoTemp.Clones = append(infoTemp.Clones, strings.Split(tmpList[clonesIndex], ",")...)
		}

		infoTemp.SizeBytes, err = strconv.ParseUint(tmpList[usedIndex], 10, 64)
		if err != nil {
			infoTemp.SizeBytes = 0
		}
		infoTemp.SizeHuman = osfreebsd.BytesToHuman(infoTemp.SizeBytes)

		info = append(info, infoTemp)
	}

	return info, nil
}

func SnapshotListWithDescriptions() ([]SnapshotInfo, error) {
	snapDescriptions := []SnapshotInfo{}
	snapList, err := SnapshotListAll()
	if err != nil {
		return []SnapshotInfo{}, err
	}

	out, err := exec.Command("zfs", "get", "-r", "hoster:sdescription").CombinedOutput()
	if err != nil {
		errString := strings.TrimSpace(string(out)) + "; " + err.Error()
		return []SnapshotInfo{}, errors.New(errString)
	}
	// Example output:
	// NAME                                               PROPERTY             VALUE                SOURCE
	// iscsihci_001/test-vm-1                             hoster:sdescription  -                    -
	// iscsihci_001/test-vm-1@custom_2023-12-25_02-22-23  hoster:sdescription  This was a test run  local
	nameIndex := -1
	propertyIndex := -1
	valueIndex := -1
	sourceIndex := -1

	// Parse the header
	reSplitSpace := regexp.MustCompile(`\s+`)
	for i, v := range strings.Split(string(out), "\n") {
		if i == 0 {
			for ii, vv := range reSplitSpace.Split(v, -1) {
				if strings.TrimSpace(vv) == "NAME" {
					nameIndex = ii
				} else if strings.TrimSpace(vv) == "PROPERTY" {
					propertyIndex = ii
				} else if strings.TrimSpace(vv) == "VALUE" {
					valueIndex = ii
				} else if strings.TrimSpace(vv) == "SOURCE" {
					sourceIndex = ii
				}
			}
		}
	}

	// Check if the header was parsed correctly
	if nameIndex == -1 {
		return []SnapshotInfo{}, errors.New("could not parse a name index")
	}
	if propertyIndex == -1 {
		return []SnapshotInfo{}, errors.New("could not parse a property index")
	}
	if valueIndex == -1 {
		return []SnapshotInfo{}, errors.New("could not parse a value index")
	}
	if sourceIndex == -1 {
		return []SnapshotInfo{}, errors.New("could not parse a source index")
	}

	// Parse the output without a header
	for i, v := range strings.Split(string(out), "\n") {
		if i == 0 {
			continue
		}
		if len(v) < 1 {
			continue
		}

		infoTemp := SnapshotInfo{}
		tmpList := reSplitSpace.Split(v, -1)

		if len(tmpList) < 1 {
			continue
		}

		nameSplit := strings.Split(tmpList[nameIndex], "@")
		if len(nameSplit) < 2 {
			continue
		} else {
			infoTemp.Name = tmpList[nameIndex]
			infoTemp.Description = tmpList[valueIndex]
		}

		snapDescriptions = append(snapDescriptions, infoTemp)
	}

	for _, v := range snapDescriptions {
		for ii, vv := range snapList {
			if v.Name == vv.Name {
				snapList[ii].Description = v.Description
			} else {
				continue
			}
		}
	}

	return snapList, nil
}

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

func RemoveSnapshot(snapshotName string) error {
	reMatchAt := regexp.MustCompile("@")
	if !reMatchAt.MatchString(snapshotName) {
		return errors.New("cannot remove, not a snapshot")
	}

	out, err := exec.Command("zfs", "destroy", snapshotName).CombinedOutput()
	if err != nil {
		errString := strings.TrimSpace(string(out)) + "; " + err.Error()
		return errors.New(errString)
	}

	return nil
}

// Finds a dataset for any given VM or a Jail, and returns it's ZFS name as a string
func FindResourceDataset(resName string) (string, error) {
	dsList, err := DefaultDatasetList()
	if err != nil {
		return "", err
	}

	reMatchName := regexp.MustCompile(`/` + resName + "$")
	resFound := false
	dsName := ""
	for _, v := range dsList {
		if reMatchName.MatchString(v) {
			dsName = v
			resFound = true
			break
		}
	}

	if !resFound {
		return "", errors.New("resource was not found")
	}

	return dsName, nil
}

// Simply returns a list of available ZFS datasets, using a default ZFS list command
func DefaultDatasetList() ([]string, error) {
	out, err := exec.Command("zfs", "list", "-p").CombinedOutput()
	if err != nil {
		errString := strings.TrimSpace(string(out)) + "; " + err.Error()
		return []string{}, errors.New(errString)
	}

	reSplitSpace := regexp.MustCompile(`\s+`)
	result := []string{}
	for i, v := range strings.Split(string(out), "\n") {
		if i == 0 {
			continue
		}

		v = reSplitSpace.Split(v, -1)[0]
		v = strings.TrimSpace(v)
		result = append(result, v)
	}

	return result, nil
}
