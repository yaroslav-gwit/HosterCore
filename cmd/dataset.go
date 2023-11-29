package cmd

import (
	"errors"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	"github.com/aquasecurity/table"
	"github.com/spf13/cobra"
)

var (
	datasetCmd = &cobra.Command{
		Use:   "dataset",
		Short: "ZFS Dataset related operations",
		Long:  `ZFS Dataset related operations.`,
		Run: func(cmd *cobra.Command, args []string) {
			checkInitFile()
			cmd.Help()
		},
	}
)

var (
	datasetListUnixStyleTable bool

	datasetListCmd = &cobra.Command{
		Use:   "list",
		Short: "List all Hoster related ZFS datasets",
		Long:  `List all Hoster related ZFS datasets.`,
		Run: func(cmd *cobra.Command, args []string) {
			checkInitFile()
			printZfsDatasetInfo()
		},
	}
)

type ZfsDatasetInfo struct {
	Name           string
	MountPoint     string
	SpaceFree      int
	SpaceFreeHuman string
	SpaceUsed      int
	SpaceUsedHuman string
	Encrypted      bool
}

func printZfsDatasetInfo() {
	dsInfo, err := getZfsDatasetInfo()
	if err != nil {
		log.Fatal(err)
	}

	var ID = 0
	var t = table.New(os.Stdout)
	t.SetAlignment(table.AlignRight, //ID
		table.AlignLeft,  // ZFS Dataset
		table.AlignRight, // Space Used
		table.AlignRight, // Space Free
		table.AlignRight, // Encrypted
	)

	if datasetListUnixStyleTable {
		t.SetDividers(table.Dividers{
			ALL: " ",
			NES: " ",
			NSW: " ",
			NEW: " ",
			ESW: " ",
			NE:  " ",
			NW:  " ",
			SW:  " ",
			ES:  " ",
			EW:  " ",
			NS:  " ",
		})
		t.SetRowLines(false)
		t.SetBorderTop(false)
		t.SetBorderBottom(false)
	} else {
		t.SetHeaders("Hoster ZFS Datasets")
		t.SetHeaderColSpans(0, 5)

		t.AddHeaders(
			"#",
			"ZFS Dataset",
			"Space Used",
			"Space Available",
			"Encrypted",
		)

		t.SetLineStyle(table.StyleBrightCyan)
		t.SetDividers(table.UnicodeRoundedDividers)
		t.SetHeaderStyle(table.StyleBold)
	}

	for _, v := range dsInfo {
		ID = ID + 1

		encrypted := "no"
		if v.Encrypted {
			encrypted = "yes"
		}
		t.AddRow(
			strconv.Itoa(ID),
			v.Name,
			v.SpaceUsedHuman,
			v.SpaceFreeHuman,
			encrypted,
		)
	}

	t.Render()
}

func getZfsDatasetInfo() ([]ZfsDatasetInfo, error) {
	zfsDatasetInfo := []ZfsDatasetInfo{}
	hostInfo, err := GetHostConfig()
	if err != nil {
		return []ZfsDatasetInfo{}, err
	}

	out, err := exec.Command("zfs", "list", "-Hp").CombinedOutput()
	// Example output:
	//     	[0]     				[1]				[2]  		 [3] 			[4]
	// zroot/vm-encrypted      205033119744    769681932288    425984  /zroot/vm-encrypted
	// zroot/vm-unencrypted    98304           769681932288    98304   /zroot/vm-unencrypted

	if err != nil {
		return []ZfsDatasetInfo{}, errors.New("output: " + string(out) + " " + err.Error())
	}

	reSplitSpace := regexp.MustCompile(`\s+`)
	for _, v := range strings.Split(string(out), "\n") {
		cmdSplitLine := reSplitSpace.Split(v, -1)
		for _, vv := range hostInfo.ActiveDatasets {
			if vv == cmdSplitLine[0] {
				tempZfsDs := ZfsDatasetInfo{}
				tempZfsDs.Name = cmdSplitLine[0]
				tempZfsDs.SpaceUsed, _ = strconv.Atoi(cmdSplitLine[1])
				tempZfsDs.SpaceUsedHuman = ByteConversion(tempZfsDs.SpaceUsed)
				tempZfsDs.SpaceFree, _ = strconv.Atoi(cmdSplitLine[2])
				tempZfsDs.SpaceFreeHuman = ByteConversion(tempZfsDs.SpaceFree)
				tempZfsDs.MountPoint = cmdSplitLine[4]
				zfsDatasetInfo = append(zfsDatasetInfo, tempZfsDs)
			}
		}
	}

	for i, v := range zfsDatasetInfo {
		out, err := exec.Command("zfs", "get", "-H", "encryption", v.Name).CombinedOutput()
		// Example output:
		//      [0]                    [1]          [2]       [3]
		// zroot/vm-unencrypted    encryption      off     default

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
