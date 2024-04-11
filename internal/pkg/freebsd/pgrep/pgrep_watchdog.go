package FreeBSDPgrep

import (
	"fmt"
	"regexp"
	"strings"
)

func FindWatchdog() (r int, e error) {
	label := "ha_watchdog"

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
