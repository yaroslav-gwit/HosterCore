package zfsutils

import (
	"HosterCore/osfreebsd"
	"errors"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

type SnapshotInfo struct {
	Dataset   string   `json:"snapshot_dataset"`
	Name      string   `json:"snapshot_name"`
	ShortName string   `json:"snapshot_short_name"`
	Locked    bool     `json:"snapshot_locked"`
	Clones    []string `json:"snapshot_clones"`
	SizeBytes uint64   `json:"snapshot_size_bytes"`
	SizeHuman string   `json:"snapshot_size_human"`
}

// Returns all ZFS snapshots present on this system
func SnapshotList() ([]SnapshotInfo, error) {
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
