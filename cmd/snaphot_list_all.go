package cmd

import (
	"HosterCore/emojlog"
	"HosterCore/zfsutils"
	"fmt"
	"os"
	"regexp"
	"strconv"

	"github.com/aquasecurity/table"
	"github.com/spf13/cobra"
)

var (
	snapshotListAllUnixStyleTable bool

	snapshotListAllCmd = &cobra.Command{
		Use:   "list-all",
		Short: "List all ZFS snapshots",
		Long:  `List all ZFS snapshots that are a part of Hoster VMs`,
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			checkInitFile()
			err := generateSnapshotAllTable()
			if err != nil {
				emojlog.PrintLogMessage(err.Error(), emojlog.Error)
				os.Exit(1)
			}
		},
	}
)

func generateSnapshotAllTable() error {
	var ID = 0
	var t = table.New(os.Stdout)
	t.SetAlignment(table.AlignRight, //ID
		table.AlignLeft,   // Resource Name
		table.AlignCenter, // Resource Type
		table.AlignLeft,   // Snapshot Name
		table.AlignRight,  // Snapshot Size Human
		table.AlignRight,  // Snapshot Size Bytes
		table.AlignRight,  // Snapshot Locked
		table.AlignRight,  // Snapshot Clones/Dependents
		table.AlignRight,  // Snapshot Description
	)

	if snapshotListAllUnixStyleTable {
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
		t.SetHeaderColSpans(0, 8)

		t.AddHeaders(
			"#",
			"Resource\nName",
			"Resource\nType",
			"Snapshot Name",
			"Snapshot\nSize Human",
			"Snapshot\nSize Bytes",
			"Snapshot\nLocked",
			"Snapshot\nDependents",
			"Snapshot\nDescription",
		)

		t.SetLineStyle(table.StyleBrightCyan)
		t.SetDividers(table.UnicodeRoundedDividers)
		t.SetHeaderStyle(table.StyleBold)
	}

	vmList := getAllVms()
	jailList, _ := GetAllJailsList()
	snapList, err := zfsutils.SnapshotListWithDescriptions()
	if err != nil {
		return err
	}

	for _, v := range vmList {
		reMatch := regexp.MustCompile(`/` + v + `@`)
		for _, vv := range snapList {
			if reMatch.MatchString(vv.Name) {
				ID = ID + 1
				t.AddRow(
					strconv.Itoa(ID),
					v,
					"VM",
					vv.Name,
					vv.SizeHuman,
					fmt.Sprintf("%d", vv.SizeBytes),
					fmt.Sprintf("%v", vv.Locked),
					fmt.Sprintf("%d", len(vv.Clones)),
					vv.Description,
				)
			}
		}
	}
	for _, v := range jailList {
		reMatch := regexp.MustCompile(`/` + v + `@`)
		for _, vv := range snapList {
			if reMatch.MatchString(vv.Name) {
				ID = ID + 1
				t.AddRow(
					strconv.Itoa(ID),
					v,
					"Jail",
					vv.Name,
					vv.SizeHuman,
					fmt.Sprintf("%d", vv.SizeBytes),
					fmt.Sprintf("%v", vv.Locked),
					fmt.Sprintf("%d", len(vv.Clones)),
					vv.Description,
				)
			}
		}
	}

	// for _, vm := range getAllVms() {
	// 	info, err := GetSnapshotInfo(vm, true)
	// 	if err != nil {
	// 		return err
	// 	}

	// 	for _, vmSnap := range info {
	// 		ID = ID + 1
	// 		t.AddRow(strconv.Itoa(ID),
	// 			vm,
	// 			vmSnap.Name,
	// 			vmSnap.SizeHuman,
	// 			strconv.Itoa(int(vmSnap.SizeBytes)),
	// 		)
	// 	}
	// }

	t.Render()
	return nil
}
