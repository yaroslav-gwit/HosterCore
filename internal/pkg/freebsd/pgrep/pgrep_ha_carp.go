// Copyright 2024 Hoster Authors. All rights reserved.
// Use of this source code is governed by an Apache License 2.0
// license that can be found in the LICENSE file.

package FreeBSDPgrep

import (
	"fmt"
	"regexp"
	"strings"
)

func FindHaCarp() (r int, e error) {
	label := "ha_carp"

	pids, err := Pgrep(label)
	if err != nil {
		e = fmt.Errorf("%s IS NOT running", label)
		return
	}
	if len(pids) < 1 {
		e = fmt.Errorf("%s IS NOT running", label)
		return
	}

	reMatchScheduler := regexp.MustCompile(`/` + label + `$`)
	for _, v := range pids {
		temp := strings.TrimSpace(v.ProcessCmd)
		if reMatchScheduler.MatchString(temp) {
			r = v.ProcessId
			return
		}
	}

	e = fmt.Errorf("%s IS NOT running", label)
	return
}
