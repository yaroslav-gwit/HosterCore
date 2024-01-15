// Copyright 2023 Hoster Authors. All rights reserved.
// Use of this source code is governed by an Apache License 2.0
// license that can be found in the LICENSE file.

package zfsutils

import (
	"errors"
	"os/exec"
	"regexp"
	"strings"
)

func SnapshotListWithDescriptions() ([]SnapshotInfo, error) {
	snapDescriptions := []SnapshotInfo{}
	snapList, err := SnapshotListAll()
	if err != nil {
		return []SnapshotInfo{}, err
	}

	out, err := exec.Command("zfs", "get", "-o", "name,property,value", "-r", "hoster:sdescription").CombinedOutput()
	if err != nil {
		errString := strings.TrimSpace(string(out)) + "; " + err.Error()
		return []SnapshotInfo{}, errors.New(errString)
	}
	// Example output:
	// NAME                                               PROPERTY             VALUE
	// iscsihci_001/test-vm-1                             hoster:sdescription  -
	// iscsihci_001/test-vm-1@custom_2023-12-25_02-22-23  hoster:sdescription  This was a test run
	nameIndex := -1
	propertyIndex := -1
	valueIndex := -1

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
			descriptionList := tmpList[valueIndex : len(tmpList)-1]
			descriptionString := strings.Join(descriptionList, " ")
			descriptionString = strings.TrimSpace(descriptionString)
			if len(descriptionString) > 0 {
				infoTemp.Description = descriptionString
			} else {
				infoTemp.Description = "-"
			}
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
