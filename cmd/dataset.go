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
			err := checkInitFile()
			if err != nil {
				log.Fatal(err.Error())
			}
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
			err := checkInitFile()
			if err != nil {
				log.Fatal(err.Error())
			}
			printZfsDatasetInfo()
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

func printZfsDatasetInfo() {
	dsInfo, err := getZfsDatasetInfo()
	if err != nil {
		log.Fatal(err)
	}

	var ID = 0
	var t = table.New(os.Stdout)
	t.SetAlignment(table.AlignRight, //ID
		table.AlignLeft, // ZFS Dataset
		table.AlignLeft, // Space Used
		table.AlignLeft, // Space Free
		table.AlignLeft, // Encrypted
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
		t.SetHeaderColSpans(0, 7)

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
