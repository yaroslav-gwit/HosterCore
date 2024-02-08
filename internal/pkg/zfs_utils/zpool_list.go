// Copyright 2023 Hoster Authors. All rights reserved.
// Use of this source code is governed by an Apache License 2.0
// license that can be found in the LICENSE file.

package zfsutils

import (
	"HosterCore/internal/pkg/byteconversion"
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

type ZpoolInfo struct {
	Healthy        bool   `json:"healthy"`
	Fragmentation  int    `json:"fragmentation"`
	Name           string `json:"name"`
	SizeHuman      string `json:"size_human"`
	SizeBytes      uint64 `json:"size_bytes"`
	AllocatedHuman string `json:"allocated_human"`
	AllocatedBytes uint64 `json:"allocated_bytes"`
	FreeHuman      string `json:"free_human"`
	FreeBytes      uint64 `json:"free_bytes"`
}

func GetZpoolList() (r []ZpoolInfo, e error) {
	out, err := exec.Command("zpool", "list", "-p").CombinedOutput()
	if err != nil {
		e = fmt.Errorf("%s; %s", strings.TrimSpace(string(out)), err.Error())
		return
	}
	// Example output
	//  [0]             [1]           [2]            [3]    [4]       [5]      [6]     [7]   [8]     [9]      [10]
	// NAME            SIZE         ALLOC           FREE  CKPOINT  EXPANDSZ   FRAG    CAP  DEDUP    HEALTH  ALTROOT
	// tank   1992864825344  563004203008  1429860622336        -         -      4     28   1.00    ONLINE  -
	// zroot    28454158336    5661163520    22792994816        -         -      5     19   1.00    ONLINE  -

	reSplitSpace := regexp.MustCompile(`\s+`)
	for i, v := range strings.Split(string(out), "\n") {
		if len(strings.TrimSpace(v)) < 1 {
			continue
		}
		if i == 0 {
			continue
		}

		pool := ZpoolInfo{}
		split := reSplitSpace.Split(v, -1)
		pool.Name = split[0]

		size, err := strconv.ParseUint(split[1], 10, 64)
		if err != nil {
			size = 0
		}
		pool.SizeBytes = size
		pool.SizeHuman = byteconversion.BytesToHuman(size)

		alloc, err := strconv.ParseUint(split[2], 10, 64)
		if err != nil {
			alloc = 0
		}
		pool.AllocatedBytes = alloc
		pool.AllocatedHuman = byteconversion.BytesToHuman(alloc)

		free, err := strconv.ParseUint(split[3], 10, 64)
		if err != nil {
			free = 0
		}
		pool.FreeBytes = free
		pool.FreeHuman = byteconversion.BytesToHuman(free)

		frag, err := strconv.Atoi(split[6])
		if err != nil {
			frag = 0
		}
		pool.Fragmentation = frag

		if split[9] == "ONLINE" {
			pool.Healthy = true
		}

		r = append(r, pool)
	}

	return
}
