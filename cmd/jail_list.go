package cmd

import (
	"HosterCore/internal/pkg/emojlog"
	HosterJail "HosterCore/internal/pkg/hoster/jail"
	"fmt"
	"os"

	"github.com/aquasecurity/table"
	"github.com/spf13/cobra"
)

var (
	jailListCmdUnixStyle bool

	jailListCmd = &cobra.Command{
		Use:   "list",
		Short: "List all available Jails in a single table",
		Long:  `List all available Jails in a single table.`,
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			checkInitFile()

			err := generateJailsTable(jailListCmdUnixStyle)
			if err != nil {
				emojlog.PrintLogMessage(err.Error(), emojlog.Error)
				os.Exit(1)
			}
		},
	}
)

func generateJailsTable(unixStyleTable bool) error {
	jailList, err := HosterJail.ListAllExtendedTable()
	if err != nil {
		return err
	}

	var t = table.New(os.Stdout)
	t.SetAlignment(table.AlignRight, //ID
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

	ID := 0
	for _, v := range jailList {
		ID += 1
		t.AddRow(fmt.Sprintf("%d", ID),
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
