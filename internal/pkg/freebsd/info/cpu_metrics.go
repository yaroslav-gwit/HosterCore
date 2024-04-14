package FreeBSDOsInfo

import (
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

type VmStatCpu struct {
	UserMode   int `json:"user_mode"`   // % of cpu time in user mode
	SystemMode int `json:"system_mode"` // % of cpu time in system mode
	IdleMode   int `json:"idle_mode"`   // % of cpu time in idle mode
}

type IoStatCpu struct {
	UserMode            int `json:"user_mode"`             // % of cpu time in user mode
	UserNiceMode        int `json:"user_nice_mode"`        // % of cpu time in user nice mode
	SystemMode          int `json:"system_mode"`           // % of cpu time in system mode
	SystemInterruptMode int `json:"system_interrupt_mode"` // % of cpu time in system interrupt mode
	IdleMode            int `json:"idle_mode"`             // % of cpu time in idle mode
}

func VmStatCpuMetrics() (r []VmStatCpu, e error) {
	out, err := exec.Command("vmstat", "-P").CombinedOutput()
	if err != nil {
		e = err
		return
	}
	outTrimmed := strings.TrimSpace(string(out))
	splitSpace := regexp.MustCompile(`\s+`)

	userModeIndexes := []int{}
	for i, v := range strings.Split(outTrimmed, "\n") {
		if i == 0 {
			continue
		}

		if i == 1 {
			for ii, vv := range splitSpace.Split(v, -1) {
				if strings.TrimSpace(vv) == "us" {
					userModeIndexes = append(userModeIndexes, ii)
				}
			}
		}

		if i == 2 {
			for _, vv := range userModeIndexes {
				temp := VmStatCpu{}
				valueList := splitSpace.Split(v, -1)
				if len(valueList) < (vv + 4) {
					e = fmt.Errorf("could not get all values")
					return
				}

				userMode, err := strconv.Atoi(valueList[vv+1])
				if err != nil {
					e = fmt.Errorf("could not get user mode usage: %s", err.Error())
					return
				}

				systemMode, err := strconv.Atoi(valueList[vv+2])
				if err != nil {
					e = fmt.Errorf("could not get system mode usage: %s", err.Error())
					return
				}

				idleMode, err := strconv.Atoi(valueList[vv+3])
				if err != nil {
					e = fmt.Errorf("could not get idle mode %%: %s", err.Error())
					return
				}

				temp.UserMode = userMode
				temp.SystemMode = systemMode
				temp.IdleMode = idleMode
				r = append(r, temp)
			}
		}
	}

	return
}

func IoStatCpuMetrics() (r IoStatCpu, e error) {
	out, err := exec.Command("iostat").CombinedOutput()
	if err != nil {
		e = err
		return
	}
	outTrimmed := strings.TrimSpace(string(out))
	splitSpace := regexp.MustCompile(`\s+`)
	userModeIndex := -1

	for i, v := range strings.Split(outTrimmed, "\n") {
		if i == 0 {
			continue
		}

		if i == 1 {
			for ii, vv := range splitSpace.Split(v, -1) {
				if strings.TrimSpace(vv) == "us" {
					userModeIndex = ii
				}
			}
		}

		if userModeIndex < 0 {
			e = fmt.Errorf("could not find user mode index")
			return
		}

		if i == 2 {
			values := splitSpace.Split(v, -1)
			if len(values) < (userModeIndex + 5) {
				break
			}

			r.UserMode, err = strconv.Atoi(values[userModeIndex])
			if err != nil {
				e = fmt.Errorf("could not get user mode usage: %s", err.Error())
				return
			}

			r.UserNiceMode, err = strconv.Atoi(values[userModeIndex+1])
			if err != nil {
				e = fmt.Errorf("could not get user nice mode usage: %s", err.Error())
				return
			}

			r.SystemMode, err = strconv.Atoi(values[userModeIndex+2])
			if err != nil {
				e = fmt.Errorf("could not get system mode usage: %s", err.Error())
				return
			}

			r.SystemInterruptMode, err = strconv.Atoi(values[userModeIndex+3])
			if err != nil {
				e = fmt.Errorf("could not get system interrupt mode usage: %s", err.Error())
				return
			}

			r.IdleMode, err = strconv.Atoi(values[userModeIndex+4])
			if err != nil {
				e = fmt.Errorf("could not get idle mode %%: %s", err.Error())
				return
			}
		}
	}

	return
}
