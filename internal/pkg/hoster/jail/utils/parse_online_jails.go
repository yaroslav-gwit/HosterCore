package HosterJailUtils

import (
	"errors"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

type LiveJailStruct struct {
	ID         int
	Name       string
	Path       string
	Dying      bool
	Ip4address string
	Ip6address string
}

// Gets the list of actively running Jails using the underlying `jls` command.
//
// Only used to compare with the list of deployed Jails, to figure out if the Jail is running (live) or not.
func GetRunningJails() ([]LiveJailStruct, error) {
	jails := []LiveJailStruct{}

	out, err := exec.Command("jls", "-h", "jid", "name", "path", "dying", "ip4.addr", "ip6.addr").CombinedOutput()
	// Example output this implementation is based on
	// jid  name   path        dying   ip4.addr   ip6.addr
	// [0]  [1]     [2]         [3]       [4]      [5]
	// 1  example /root/jail   false  10.0.105.50   -
	// 2  twelve  /root/12_4   false  10.0.105.51   -
	// 3  twelve1 /root/12_4_1 false  10.0.105.52   -
	// 4  twelve2 /root/12_4_2 false  10.0.105.53   -
	// 5  twelve3 /root/12_4_3 false  10.0.105.54   -

	if err != nil {
		errorValue := strings.TrimSpace(string(out)) + "; " + err.Error()
		return []LiveJailStruct{}, errors.New(errorValue)
	}

	reSpaceSplit := regexp.MustCompile(`\s+`)
	for i, v := range strings.Split(string(out), "\n") {
		// Skip the header
		if i == 0 {
			continue
		}
		// Skip empty lines
		if len(v) < 1 {
			continue
		}

		tempList := reSpaceSplit.Split(strings.TrimSpace(v), -1)
		// In case we need to check the split output in the future
		// fmt.Println(tempList)

		tempStruct := LiveJailStruct{}
		jailId, err := strconv.Atoi(tempList[0])
		if err != nil {
			return []LiveJailStruct{}, err
		}

		tempStruct.ID = jailId
		tempStruct.Name = tempList[1]
		tempStruct.Path = tempList[2]

		tempStruct.Dying, err = strconv.ParseBool(tempList[3])
		if err != nil {
			return []LiveJailStruct{}, err
		}

		tempStruct.Ip4address = tempList[4]
		tempStruct.Ip6address = tempList[5]

		jails = append(jails, tempStruct)
	}

	return jails, nil
}
