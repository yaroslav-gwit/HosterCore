// Copyright 2023 Hoster Authors. All rights reserved.
// Use of this source code is governed by an Apache License 2.0
// license that can be found in the LICENSE file.

package HosterVmUtils

import (
	"HosterCore/internal/pkg/byteconversion"
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

func DiskInfo(filePath string) (r DiskInfoApi, e error) {
	reSplitSpace := regexp.MustCompile(`\s+`)

	// Total Disk Space
	out, err := exec.Command("ls", "-al", filePath).CombinedOutput()
	if err != nil {
		e = fmt.Errorf("%s; %s", strings.TrimSpace(string(out)), err.Error())
		return
	}
	// Example output
	//    [0]     [1] [2]   [3]      [4]        ...doesn't matter
	// -rw-------  1 root  wheel  5368709120 Jan 12 00:35 /tank/vm-encrypted/test-vm-1/disk0.img

	temp := strings.TrimSpace(string(out))
	split := reSplitSpace.Split(temp, -1)

	r.TotalBytes, err = strconv.ParseUint(split[4], 10, 64)
	if err != nil {
		e = err
		return
	}
	r.TotalHuman = byteconversion.BytesToHuman(r.TotalBytes)
	// EOF Total Disk Space

	// Free Disk Space
	out, err = exec.Command("du", filePath).CombinedOutput()
	if err != nil {
		e = fmt.Errorf("%s; %s", strings.TrimSpace(string(out)), err.Error())
		return
	}

	temp = strings.TrimSpace(string(out))
	split = reSplitSpace.Split(temp, -1)

	r.UsedBytes, err = strconv.ParseUint(split[0], 10, 64)
	if err != nil {
		e = err
		return
	}
	r.UsedBytes = r.UsedBytes * 1024
	r.UsedHuman = byteconversion.BytesToHuman(r.UsedBytes)
	// EOF Free Disk Space

	return
}
