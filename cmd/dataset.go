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
	datasetCmd = &cobra.Command{
		Use:   "dataset",
		Short: "ZFS Dataset related operations",
		Long:  `ZFS Dataset related operations.`,
		Run: func(cmd *cobra.Command, args []string) {
			err := checkInitFile()
			if err != nil {
				log.Fatal(err.Error())
			}
			// cmd.Help()
			fmt.Println(getZfsDatasetInfo())
		},
	}
)

type ZfsDatasetInfo struct {
	Name           string
	SpaceFree      int
	SpaceFreeHuman string
	SpaceUsed      int
	SpaceUsedHuman string
	Encrypted      bool
}

func getZfsDatasetInfo() ([]ZfsDatasetInfo, error) {
	zfsDatasetInfo := []ZfsDatasetInfo{}
	hostInfo, err := GetHostConfig()
	if err != nil {
		return []ZfsDatasetInfo{}, err
	}

	// Standard command output:
	// zroot/vm-encrypted      205033119744    769681932288    425984  /zroot/vm-encrypted
	// zroot/vm-unencrypted    98304   769681932288    98304   /zroot/vm-unencrypted
	//
	out, err := exec.Command("zfs", "list", "-Hp").CombinedOutput()
	if err != nil {
		return []ZfsDatasetInfo{}, errors.New("Output: " + string(out) + " Status code: " + err.Error())
	}

	reSplitSpace := regexp.MustCompile(`\s+`)
	for _, v := range strings.Split(string(out), "\n") {
		temp := reSplitSpace.Split(v, -1)
		for _, vv := range hostInfo.ActiveDatasets {
			if vv == temp[0] {
				tempZfsDs := ZfsDatasetInfo{}
				tempZfsDs.Name = temp[0]
				tempZfsDs.SpaceFree, _ = strconv.Atoi(temp[1])
				tempZfsDs.SpaceFreeHuman = ByteConversion(tempZfsDs.SpaceFree)
				tempZfsDs.SpaceUsed, _ = strconv.Atoi(temp[2])
				tempZfsDs.SpaceUsedHuman = ByteConversion(tempZfsDs.SpaceUsed)
				zfsDatasetInfo = append(zfsDatasetInfo, tempZfsDs)
			}
		}
	}

	for i, v := range zfsDatasetInfo {
		// Standard command output:
		// zroot/vm-unencrypted    encryption      off     default
		//
		out, err := exec.Command("zfs", "get", "-H", "encryption", v.Name).CombinedOutput()
		if err != nil {
			return []ZfsDatasetInfo{}, errors.New("Output: " + string(out) + " Status code: " + err.Error())
		}

		zfsDatasetInfo[i].Encrypted = false
		for ii, vv := range reSplitSpace.Split(string(out), -1) {
			if ii != 2 {
				continue
			}
			if vv != "off" {
				zfsDatasetInfo[i].Encrypted = true
			}
		}
	}

	return zfsDatasetInfo, nil
}
