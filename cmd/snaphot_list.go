//go:build freebsd
// +build freebsd

package cmd

import (
	"HosterCore/internal/pkg/byteconversion"
	"HosterCore/internal/pkg/emojlog"
	HosterJailUtils "HosterCore/internal/pkg/hoster/jail/utils"
	HosterVmUtils "HosterCore/internal/pkg/hoster/vm/utils"
	zfsutils "HosterCore/internal/pkg/zfs_utils"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	"github.com/aquasecurity/table"
	"github.com/spf13/cobra"
)

var (
	snapshotListUnixStyleTable bool

	snapshotListCmd = &cobra.Command{
		Use:   "list [vmName or jailName]",
		Short: "List VM specific snapshots",
		Long:  `List VM specific snapshot information including snapshot name, size and time taken`,
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			checkInitFile()

			err := generateSnapshotTableNew(args[0])
			if err != nil {
				emojlog.PrintLogMessage(err.Error(), emojlog.Error)
				os.Exit(1)
			}
		},
	}
)

type SnapshotInfo struct {
	Name      string `json:"snapshot_name"`
	SizeBytes uint64 `json:"snapshot_size_bytes"`
	SizeHuman string `json:"snapshot_size_human"`
}

// Returns a list of ZFS snapshots for a particular VM, along with other useful information,
// like snapshot size in bytes, and human readable snapshot size
func GetSnapshotInfo(vmName string, ignoreVmExistsCheck bool) ([]SnapshotInfo, error) {
	if !ignoreVmExistsCheck {
		vmList, _ := HosterVmUtils.ListAllSimple()
		vmExists := false
		for _, v := range vmList {
			if v.VmName == vmName {
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
		tempInfo.SizeHuman = byteconversion.BytesToHuman(tempInfo.SizeBytes)
		snapshotInfo = append(snapshotInfo, tempInfo)
	}

	return snapshotInfo, nil
}

func generateSnapshotTableNew(vmName string) error {
	var ID = 0
	var t = table.New(os.Stdout)
	t.SetAlignment(table.AlignRight, //ID
		table.AlignLeft,   // Resource Name
		table.AlignCenter, // Resource Type
		table.AlignLeft,   // Snapshot Name
		table.AlignRight,  // Snapshot Size Human
		// table.AlignRight,  // Snapshot Size Bytes
		table.AlignCenter, // Snapshot Locked
		table.AlignRight,  // Snapshot Clones/Dependents
		table.AlignRight,  // Snapshot Description
	)

	if snapshotListUnixStyleTable {
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
		t.SetHeaders("Hoster ZFS Snapshots")
		// t.SetHeaderColSpans(0, 9)
		t.SetHeaderColSpans(0, 8)

		t.AddHeaders(
			"#",
			"Resource\nName",
			"Resource\nType",
			"Snapshot\nName",
			"Snapshot Size\nHuman",
			// "Snapshot Size\nBytes",
			"Snapshot\nLocked",
			"Snapshot\nDependents",
			"Snapshot\nDescription",
		)

		t.SetLineStyle(table.StyleBrightCyan)
		t.SetDividers(table.UnicodeRoundedDividers)
		t.SetHeaderStyle(table.StyleBold)
	}

	resFound := false
	resType := ""
	vms, _ := HosterVmUtils.ListAllSimple()
	jails, _ := HosterJailUtils.ListAllSimple()

	for _, v := range vms {
		if v.VmName == vmName {
			resFound = true
			resType = "VM"
		}
	}
	if !resFound {
		for _, v := range jails {
			if v.JailName == vmName {
				resFound = true
				resType = "Jail"
			}
		}
	}

	if !resFound {
		return errors.New("can't find VM/Jail with this name: " + vmName)
	}

	snapList, err := zfsutils.SnapshotListWithDescriptions()
	if err != nil {
		return err
	}

	reMatch := regexp.MustCompile(`/` + vmName + `@`)
	for _, vv := range snapList {
		if reMatch.MatchString(vv.Name) {
			ID = ID + 1
			t.AddRow(
				strconv.Itoa(ID),
				vmName,
				resType,
				vv.Name,
				vv.SizeHuman,
				// fmt.Sprintf("%d", vv.SizeBytes),
				fmt.Sprintf("%v", vv.Locked),
				fmt.Sprintf("%d", len(vv.Clones)),
				vv.Description,
			)
		}
	}

	t.Render()
	return nil
}
