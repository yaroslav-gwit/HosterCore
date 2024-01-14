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
