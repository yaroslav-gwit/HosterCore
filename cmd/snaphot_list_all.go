//go:build freebsd
// +build freebsd

package cmd

import (
	"HosterCore/internal/pkg/emojlog"
	HosterJailUtils "HosterCore/internal/pkg/hoster/jail/utils"
	HosterVmUtils "HosterCore/internal/pkg/hoster/vm/utils"
	zfsutils "HosterCore/internal/pkg/zfs_utils"
	"fmt"
	"os"
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
		// table.AlignRight,  // Snapshot Size Bytes
		table.AlignCenter, // Snapshot Locked
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

	// vmList := getAllVms()
	vmList, _ := HosterVmUtils.ListAllSimple()
	jailList, _ := HosterJailUtils.ListAllSimple()
	snapList, err := zfsutils.SnapshotListWithDescriptions()
	if err != nil {
		return err
	}

	for _, v := range vmList {
		for _, vv := range snapList {
			if vv.Dataset == v.DsName+"/"+v.VmName {
				ID = ID + 1
				t.AddRow(
					strconv.Itoa(ID),
					v.VmName,
					"VM",
					vv.Name,
					vv.SizeHuman,
					fmt.Sprintf("%v", vv.Locked),
					fmt.Sprintf("%d", len(vv.Clones)),
					vv.Description,
				)
			}
		}
	}

	for _, v := range jailList {
		for _, vv := range snapList {
			if vv.Dataset == v.DsName+"/"+v.JailName {
				ID = ID + 1
				t.AddRow(
					strconv.Itoa(ID),
					v.JailName,
					"Jail",
					vv.Name,
					vv.SizeHuman,
					fmt.Sprintf("%v", vv.Locked),
					fmt.Sprintf("%d", len(vv.Clones)),
					vv.Description,
				)
			}
		}
	}

	t.Render()
	return nil
}
