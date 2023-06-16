package cmd

import (
	"errors"
	"fmt"
	"log"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

var (
	vmSnapshotListCmd = &cobra.Command{
		Use:   "snapshot-list [vmName]",
		Short: "List VM specific snapshots",
		Long:  `List VM specific snapshot information including snapshot name, size and time taken`,
		Run: func(cmd *cobra.Command, args []string) {
			err := checkInitFile()
			if err != nil {
				log.Fatal(err.Error())
			}
			info, err := getSnapshotInfo(args[0])
			if err != nil {
				log.Fatal(err)
			}
			fmt.Println(info)
		},
	}
)

type SnapshotInfo struct {
	Name      string
	SizeBytes uint64
	SizeHuman string
}

func getSnapshotInfo(vmName string) ([]SnapshotInfo, error) {
	vmList := getAllVms()
	vmExists := false
	for _, v := range vmList {
		if v == vmName {
			vmExists = true
		}
	}

	if !vmExists {
		return []SnapshotInfo{}, errors.New("vm was not found")
	}

	snapshotInfo := []SnapshotInfo{}
	vmDataset, err := getVmDataset(vmName)
	if err != nil {
		return []SnapshotInfo{}, err
	}

	out, err := exec.Command("zfs", "list", "-rpH", "-t", "snapshot", "-o", "name,used", vmDataset).CombinedOutput()
	if err != nil {
		fmt.Println(string(out))
		return []SnapshotInfo{}, err
	}

	reSplitSpace := regexp.MustCompile(`\s+`)
	for _, v := range strings.Split(string(out), "\n") {
		if len(v) < 1 {
			continue
		}
		tempList := reSplitSpace.Split(v, -1)
		tempInfo := SnapshotInfo{}
		tempInfo.Name = tempList[0]
		tempInfo.SizeBytes, _ = strconv.ParseUint(tempList[1], 10, 64)
		tempInfo.SizeHuman = ByteConversion(int(tempInfo.SizeBytes))
		snapshotInfo = append(snapshotInfo, tempInfo)
	}

	return snapshotInfo, nil
}
