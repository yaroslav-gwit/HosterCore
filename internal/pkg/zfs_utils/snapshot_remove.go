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
