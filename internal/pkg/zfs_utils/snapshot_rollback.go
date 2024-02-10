// Copyright 2023 Hoster Authors. All rights reserved.
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

func RollbackSnapshot(snapshotName string) error {
	reMatchAt := regexp.MustCompile("@")
	if !reMatchAt.MatchString(snapshotName) {
		return errors.New("not a snapshot, provide a correct snapshot name")
	}

	out, err := exec.Command("zfs", "rollback", "-r", snapshotName).CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s; %s", strings.TrimSpace(string(out)), err.Error())
	}

	return nil
}
