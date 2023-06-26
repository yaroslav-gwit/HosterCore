package cmd

import (
	"errors"
	"fmt"
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
	vmSnapshotListUnixStyleTable bool

	vmSnapshotListCmd = &cobra.Command{
		Use:   "snapshot-list [vmName]",
		Short: "List VM specific snapshots",
		Long:  `List VM specific snapshot information including snapshot name, size and time taken`,
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			err := checkInitFile()
			if err != nil {
				log.Fatal(err)
			}
			err = generateSnapshotTable(args[0])
			if err != nil {
				log.Fatal(err)
			}
		},
	}
)

func generateSnapshotTable(vmName string) error {
	info, err := getSnapshotInfo(vmName, false)
	if err != nil {
		return err
	}

	var ID = 0
	var t = table.New(os.Stdout)
	t.SetAlignment(table.AlignRight, //ID
		table.AlignLeft,  // Snapshot Name
		table.AlignRight, // Snapshot Size Human
		table.AlignRight) // Snapshot Size Bytes

	if vmSnapshotListUnixStyleTable {
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
		t.SetHeaders("List of ZFS Snapshots for: " + vmName)
		t.SetHeaderColSpans(0, 4)

		t.AddHeaders(
			"ID",
			"Snapshot Name",
			"Snapshot Size Human",
			"Snapshot Size Bytes")

		t.SetLineStyle(table.StyleBrightCyan)
		t.SetDividers(table.UnicodeRoundedDividers)
		t.SetHeaderStyle(table.StyleBold)
	}

	for _, vmSnap := range info {
		ID = ID + 1
		t.AddRow(strconv.Itoa(ID),
			vmSnap.Name,
			vmSnap.SizeHuman,
			strconv.Itoa(int(vmSnap.SizeBytes)))
	}

	t.Render()
	return nil
}

type SnapshotInfo struct {
	Name      string `json:"snapshot_name"`
	SizeBytes uint64 `json:"snapshot_size_bytes"`
	SizeHuman string `json:"snapshot_size_human"`
}

func getSnapshotInfo(vmName string, ignoreVmExistsCheck bool) ([]SnapshotInfo, error) {
	if !ignoreVmExistsCheck {
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
