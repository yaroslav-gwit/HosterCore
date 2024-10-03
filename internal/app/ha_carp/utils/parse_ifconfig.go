package CarpUtils

import (
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

// Returns an error if the ifconfig command fails, or if it cannot find any carp interfaces, or if it cannot parse the output.
func ParseIfconfig() (r []CarpInfo, e error) {
	out, err := exec.Command("ifconfig").CombinedOutput()
	if err != nil {
		e = fmt.Errorf("could not run ifconfig: %s; %s", strings.TrimSpace(string(out)), err.Error())
		return
	}

	matchString := "carp:"
	matchSpace := regexp.MustCompile(`\s+`)
	matchInterface := regexp.MustCompile(`^.*:\s+flags=`) // re0: flags=8943<UP,BROADCAST,RUNNING,PROMISC,SIMPLEX,MULTICAST> metric 0 mtu 1500

	// Example output:
	// carp: MASTER vhid 2 advbase 2 advskew 20
	currentInterface := ""
	for _, v := range strings.Split(string(out), "\n") {
		if matchInterface.MatchString(v) {
			currentInterface = strings.Split(v, ":")[0]
		}

		if strings.Contains(v, matchString) {
			v = strings.TrimSpace(v)
			carp := CarpInfo{}

			temp := matchSpace.Split(v, -1)
			if len(temp) < 7 {
				e = fmt.Errorf("could not parse carp line: %s", v)
				return
			}

			carp.Interface = currentInterface
			carp.Status = temp[1] // MASTER, BACKUP, INIT
			carp.Vhid, err = strconv.Atoi(temp[3])
			if err != nil {
				e = fmt.Errorf("could not parse vhid: %s", temp[3])
				return
			}

			carp.Advbase, err = strconv.Atoi(temp[5])
			if err != nil {
				e = fmt.Errorf("could not parse advbase: %s", temp[3])
				return
			}

			carp.Advskew, err = strconv.Atoi(temp[7])
			if err != nil {
				e = fmt.Errorf("could not parse advskew: %s", temp[3])
				return
			}

			r = append(r, carp)
		}
	}

	if len(r) < 1 {
		e = fmt.Errorf("could not find any carp interfaces")
		return
	}

	return
}
