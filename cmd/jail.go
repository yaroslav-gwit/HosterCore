package cmd

import (
	"fmt"
	"log"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

var (
	jailCmd = &cobra.Command{
		Use:   "jail",
		Short: "Jail related operations",
		Long:  `Jail related operations: deploy, stop, start, destroy, etc`,
		Run: func(cmd *cobra.Command, args []string) {
			err := checkInitFile()
			if err != nil {
				log.Fatal(err.Error())
			}

			// cmd.Help()
			fmt.Println(getRunningJails())
		},
	}
)

type LiveJailStruct struct {
	ID         int
	Name       string
	Path       string
	Running    bool
	Ip4address string
	Ip6address string
}

func getRunningJails() (jails []LiveJailStruct, returnErr error) {
	reSpaceSplit := regexp.MustCompile(`\s+`)

	out, err := exec.Command("jls", "-h", "jid", "name", "path", "dying", "ip4.addr", "ip6.addr").CombinedOutput()
	// jid name path dying ip4.addr ip6.addr
	// [0] [1]     [2]       [3]       [4]     [5]
	// 1 example /root/jail false 10.0.105.50 -
	// 2 twelve /root/12_4 false 10.0.105.51 -
	// 3 twelve1 /root/12_4_1 false 10.0.105.52 -
	// 4 twelve2 /root/12_4_2 false 10.0.105.53 -
	// 5 twelve3 /root/12_4_3 false 10.0.105.54 -

	if err != nil {
		returnErr = err
		// return
		log.Fatal("Output: "+string(out), "Exit code: "+err.Error())
	}

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
		fmt.Println(tempList)

		tempStruct := LiveJailStruct{}

		jailId, err := strconv.Atoi(tempList[0])
		if err != nil {
			returnErr = err
			return
		}

		tempStruct.ID = jailId
		tempStruct.Name = tempList[1]
		tempStruct.Path = tempList[2]
		tempStruct.Running, returnErr = strconv.ParseBool(tempList[3])
		tempStruct.Ip4address = tempList[4]
		tempStruct.Ip6address = tempList[5]
	}

	return
}
