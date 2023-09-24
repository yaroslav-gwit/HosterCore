package cmd

import (
	"log"
	"os"
	"strconv"

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
			err := checkInitFile()
			if err != nil {
				log.Fatal(err.Error())
			}
			err = generateJailsTable(jailListCmdUnixStyle)
			if err != nil {
				log.Fatal(err.Error())
			}
		},
	}
)

func generateJailsTable(unixStyleTable bool) error {
	// fmt.Println(getRunningJails())
	jailsList, err := getAllJailsList()
	if err != nil {
		return err
	}

	var ID = 0
	var t = table.New(os.Stdout)
	t.SetAlignment(table.AlignRight, //ID
		table.AlignLeft,  // Snapshot Name
		table.AlignRight, // Snapshot Size Human
		table.AlignRight) // Snapshot Size Bytes

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
		t.SetHeaderColSpans(0, 9)

		t.AddHeaders(
			"#",
			"Jail Name",
			"Jail Status",
			"CPU Limit",
			"RAM Limit",
			"Main IP Address",
			"Release",
			"Uptime",
			"Space used",
			"Jail Description")

		t.SetLineStyle(table.StyleBrightCyan)
		t.SetDividers(table.UnicodeRoundedDividers)
		t.SetHeaderStyle(table.StyleBold)
	}

	for _, v := range jailsList {
		ID = ID + 1
		t.AddRow(strconv.Itoa(ID),
			v)
	}

	t.Render()
	return nil
}
