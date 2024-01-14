package HosterZfs

import (
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

type SpaceUsedAndAvailable struct {
	Name      string // ZFS dataset name
	Used      uint64 // Dataset space used
	Available uint64 // Dataset space available
}

// Lists all ZFS datasets, and extracts some additional information.
func ListUsedAndAvailableSpace() (r []SpaceUsedAndAvailable, e error) {
	out, err := exec.Command("zfs", "list", "-p", "-o", "name,used,available").CombinedOutput()
	if err != nil {
		e = fmt.Errorf("%s; %s", strings.TrimSpace(string(out)), err.Error())
		return
	}
	// Example output
	//    [0]                                            [1]            [2]
	// NAME                                                  USED          AVAIL
	// hast_shared                                    12878303232   503188860928
	// hast_shared/jail-template-13.2-RELEASE          1455878144   503188860928
	// hast_shared/template-debian12                   5371150336   503188860928
	// tank/hast_one                                 553666740224  1868765888512
	// tank/vm-encrypted                              12765335552  1353367633920
	// tank/vm-encrypted/jail-template-13.2-RELEASE    1488908288  1353367633920
	// tank/vm-encrypted/prometheus                     504152064  1353367633920
	// tank/vm-encrypted/prometheus-hzima-0101           23560192  1353367633920

	reSplitSpace := regexp.MustCompile(`\s+`)
	for i, v := range strings.Split(string(out), "\n") {
		// skip the header
		if i == 0 {
			continue
		}
		// skip empty lines
		if len(v) < 1 {
			continue
		}

		split := reSplitSpace.Split(v, -1)
		// Guardrails from panicking, if the amount of items is less than 3
		if len(split) < 3 {
			continue
		}

		spaceUsed, err := strconv.ParseUint(split[1], 10, 64)
		if err != nil {
			e = err
			return
		}

		spaceAvailable, err := strconv.ParseUint(split[1], 10, 64)
		if err != nil {
			e = err
			return
		}

		r = append(r, SpaceUsedAndAvailable{Name: split[0], Used: spaceUsed, Available: spaceAvailable})
	}

	return
}
