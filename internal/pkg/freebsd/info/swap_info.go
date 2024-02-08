// Copyright 2023 Hoster Authors. All rights reserved.
// Use of this source code is governed by an Apache License 2.0
// license that can be found in the LICENSE file.

package FreeBSDOsInfo

import (
	"HosterCore/internal/pkg/byteconversion"
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

type SwapInfo struct {
	SwapFreeHuman    string `json:"swap_free_human"`
	SwapFreeBytes    uint64 `json:"swap_free_bytes"`
	SwapUsedHuman    string `json:"swap_used_human"`
	SwapUsedBytes    uint64 `json:"swap_used_bytes"`
	SwapOverallHuman string `json:"swap_overall_human"`
	SwapOverallBytes uint64 `json:"swap_overall_bytes"`
}

// Returns a structured SWAP information for your FreeBSD system.
func GetSwapInfo() (r SwapInfo, e error) {
	out, err := exec.Command("swapinfo").CombinedOutput()
	if err != nil {
		e = fmt.Errorf("%s; %s", strings.TrimSpace(string(out)), err)
		return
	}
	//  Example output
	//   [0]                  [1]         [2]      [3]        [4]
	// Device              1K-blocks     Used     Avail       Capacity
	// /dev/mmcsd0p3.eli   2097152        0       2097152     0%

	var swapInfoList []string
	reSplitSpace := regexp.MustCompile(`\s+`)
	for _, v := range strings.Split(string(out), "\n") {
		if len(v) < 1 {
			continue
		}
		for _, vv := range reSplitSpace.Split(v, -1) {
			if len(vv) > 0 {
				swapInfoList = append(swapInfoList, vv)
			}
		}
	}

	r.SwapOverallBytes, err = strconv.ParseUint(swapInfoList[1], 10, 64)
	if err != nil {
		r.SwapOverallBytes = 0
		r.SwapOverallHuman = "0B"
	} else {
		r.SwapOverallBytes = r.SwapOverallBytes * 1024
		r.SwapOverallHuman = byteconversion.BytesToHuman(r.SwapOverallBytes)
	}

	r.SwapUsedBytes, err = strconv.ParseUint(swapInfoList[2], 10, 64)
	if err != nil {
		r.SwapUsedBytes = 0
		r.SwapUsedHuman = "0B"
	} else {
		r.SwapUsedBytes = r.SwapUsedBytes * 1024
		r.SwapUsedHuman = byteconversion.BytesToHuman(r.SwapUsedBytes)
	}

	r.SwapFreeBytes, err = strconv.ParseUint(swapInfoList[3], 10, 64)
	if err != nil {
		r.SwapFreeBytes = 0
		r.SwapFreeHuman = "0B"
	} else {
		r.SwapFreeBytes = r.SwapFreeBytes * 1024
		r.SwapFreeHuman = byteconversion.BytesToHuman(r.SwapFreeBytes)
	}

	return
}
