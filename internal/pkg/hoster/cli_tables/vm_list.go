// Copyright 2023 Hoster Authors. All rights reserved.
// Use of this source code is governed by an Apache License 2.0
// license that can be found in the LICENSE file.

package HosterTables

import (
	HosterVmUtils "HosterCore/internal/pkg/hoster/vm/utils"
	"fmt"
	"os"

	"github.com/aquasecurity/table"
)

func GenerateVMsTable(unix bool) error {
	vms, err := HosterVmUtils.ListAllTable()
	if err != nil {
		return err
	}

	var t = table.New(os.Stdout)
	t.SetAlignment(table.AlignRight, //ID
		table.AlignLeft,   // VM Name
		table.AlignCenter, // VM Status
		table.AlignCenter, // CPU Sockets
		table.AlignCenter, // CPU Cores
		table.AlignCenter, // RAM
		table.AlignLeft,   // Main IP
		table.AlignLeft,   // OS Comment
		table.AlignLeft,   // VM Uptime
		table.AlignCenter, // OS Disk Used
		table.AlignLeft,   // Description
	)

	if unix {
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
		t.SetHeaders("Hoster VMs")
		t.SetHeaderColSpans(0, 11)

		t.AddHeaders(
			"#",
			"VM\nName",
			"VM\nStatus",
			"CPU\nSockets",
			"CPU\nCores",
			"VM\nMemory",
			"Main IP\nAddress",
			"OS\nType",
			"VM\nUptime",
			"OS Disk\n(Used/Total)",
			"VM\nDescription",
		)

		t.SetLineStyle(table.StyleBrightCyan)
		t.SetDividers(table.UnicodeRoundedDividers)
		t.SetHeaderStyle(table.StyleBold)
	}

	if unix {
		for i, v := range vms {
			t.AddRow(
				fmt.Sprintf("%d", i+1),
				v.VmName,
				v.VmStatus,
				fmt.Sprintf("%d", v.CPUSockets),
				fmt.Sprintf("%d", v.CPUCores),
				v.VmMemory,
				v.MainIpAddress,
				v.OsType,
				v.VmUptime,
				v.DiskUsedTotal,
				v.VmDescription,
			)
		}
	} else {
		for i, v := range vms {
			t.AddRow(
				fmt.Sprintf("%d", i+1),
				v.VmName,
				v.VmStatus,
				fmt.Sprintf("%d", v.CPUSockets),
				fmt.Sprintf("%d", v.CPUCores),
				v.VmMemory,
				v.MainIpAddress,
				v.OsComment,
				v.VmUptime,
				v.DiskUsedTotal,
				v.VmDescription,
			)
		}
	}

	t.Render()
	return nil
}
