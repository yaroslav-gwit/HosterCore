// Copyright 2023 Hoster Authors. All rights reserved.
// Use of this source code is governed by an Apache License 2.0
// license that can be found in the LICENSE file.

package zfsutils

import (
	"errors"
	"regexp"
)

// Finds a dataset for any given VM or a Jail, and returns it's ZFS name as a string.
//
// Uses "/" + resName + "$" as a regex to find the required resource.
//
// Returns an error, if nothing was found
func FindResourceDataset(resName string) (string, error) {
	dsList, err := DefaultDatasetList()
	if err != nil {
		return "", err
	}

	reMatchName := regexp.MustCompile(`/` + resName + `$`)
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
