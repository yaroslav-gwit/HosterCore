package cmd

import (
	"log"
	"os"
	"strconv"

	"github.com/aquasecurity/table"
	"github.com/spf13/cobra"
)

var (
	vmSnapshotListAllUnixStyleTable bool

	vmSnapshotListAllCmd = &cobra.Command{
		Use:   "snapshot-list-all",
		Short: "List all ZFS snapshots",
		Long:  `List all ZFS snapshots that are a part of Hoster VMs`,
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			err := checkInitFile()
			if err != nil {
				log.Fatal(err)
			}
			err = generateSnapshotAllTable()
			if err != nil {
				log.Fatal(err)
			}
		},
	}
)

func generateSnapshotAllTable() error {
	var ID = 0
	var t = table.New(os.Stdout)
	t.SetAlignment(table.AlignRight, //ID
		table.AlignLeft,  // Snapshot Name
		table.AlignRight, // Snapshot Size Human
		table.AlignRight) // Snapshot Size Bytes

	if vmSnapshotListAllUnixStyleTable {
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
		t.SetHeaderColSpans(0, 5)

		t.AddHeaders(
			"ID",
			"VM Name",
			"Snapshot Name",
			"Snapshot Size Human",
			"Snapshot Size Bytes")

		t.SetLineStyle(table.StyleBrightCyan)
		t.SetDividers(table.UnicodeRoundedDividers)
		t.SetHeaderStyle(table.StyleBold)
	}

	for _, vm := range getAllVms() {
		info, err := getSnapshotInfo(vm, true)
		if err != nil {
			return err
		}

		for _, vmSnap := range info {
			ID = ID + 1
			t.AddRow(strconv.Itoa(ID),
				vm,
				vmSnap.Name,
				vmSnap.SizeHuman,
				strconv.Itoa(int(vmSnap.SizeBytes)))
		}
	}

	t.Render()
	return nil
}
