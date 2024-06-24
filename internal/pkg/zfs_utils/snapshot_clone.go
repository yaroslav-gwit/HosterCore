// Copyright 2024 Hoster Authors. All rights reserved.
// Use of this source code is governed by an Apache License 2.0
// license that can be found in the LICENSE file.

package zfsutils

import (
	"errors"
	"fmt"
	"os/exec"
	"regexp"
	"strings"
)

// # This is a naive shell wrapper around the zfs clone command.
//
// @param snapshotName: The snapshot name to clone, must be in the form of "pool/dataset/oldResName@snapshot"
//
// @param newRes: The new resource name to create, must be in the form of "pool/dataset/newResourceName"
func SnapshotClone(snapshotName string, newRes string) error {
	reMatchAt := regexp.MustCompile("@")
	if !reMatchAt.MatchString(snapshotName) {
		return errors.New("not a snapshot, provide a correct snapshot name")
	}

	reMatchNotAllowed := regexp.MustCompile(`replication`)
	if reMatchNotAllowed.MatchString(snapshotName) {
		return fmt.Errorf("snapshot of type replication cannot be cloned, because it's ephemeral")
	}

	out, err := exec.Command("zfs", "clone", snapshotName, newRes).CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s; %s", strings.TrimSpace(string(out)), err.Error())
	}

	return nil
}
