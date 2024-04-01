// Copyright 2023 Hoster Authors. All rights reserved.
// Use of this source code is governed by an Apache License 2.0
// license that can be found in the LICENSE file.

package HosterTables

import (
	HosterJailUtils "HosterCore/internal/pkg/hoster/jail/utils"
	"fmt"
	"os"

	"github.com/aquasecurity/table"
)

func GenerateJailsTable(unixStyleTable bool) error {
	jailList, err := HosterJailUtils.ListAllExtendedTable()
	if err != nil {
		return err
	}

	var t = table.New(os.Stdout)
	t.SetAlignment(
		table.AlignRight,  // ID
		table.AlignLeft,   // Jail Name
		table.AlignCenter, // Jail Status
		table.AlignCenter, // CPU Limit
		table.AlignCenter, // RAM Limit
		table.AlignLeft,   // Main IP Address
		table.AlignLeft,   // Release
		table.AlignLeft,   // Uptime
		table.AlignCenter, // Space Used/Available
		table.AlignLeft,   // Description
	)

	if unixStyleTable {
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
		t.SetHeaders("Hoster Jails")
		t.SetHeaderColSpans(0, 10)

		t.AddHeaders(
			"#",
			"Jail\nName",
			"Jail\nStatus",
			"CPU\nLimit",
			"RAM\nLimit",
			"Main IP\nAddress",
			"FreeBSD\nRelease",
			"Jail\nUptime",
			"Storage\n(Used/Available)",
			"Jail\nDescription",
		)

		t.SetLineStyle(table.StyleBrightCyan)
		t.SetDividers(table.UnicodeRoundedDividers)
		t.SetHeaderStyle(table.StyleBold)
	}

	for i, v := range jailList {
		t.AddRow(
			fmt.Sprintf("%d", i+1),
			v.Name,
			v.Status,
			v.CPULimit,
			v.RAMLimit,
			v.MainIpAddress,
			v.Release,
			v.Uptime,
			v.StorageUsed+"/"+v.StorageAvailable,
			v.Description,
		)
	}

	t.Render()
	return nil
}
