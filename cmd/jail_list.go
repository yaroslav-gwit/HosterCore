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
		table.AlignLeft,   // Jail Name
		table.AlignCenter, // Jail Status
		table.AlignCenter, // CPU Limit
		table.AlignCenter, // RAM Limit
		table.AlignLeft,   // Main IP Address
		table.AlignLeft,   // Release
		table.AlignLeft,   // Uptime
		table.AlignLeft,   // Space Used
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
			"Jail Name",
			"Jail Status",
			"CPU Limit",
			"RAM Limit",
			"Main IP Address",
			"Release",
			"Uptime",
			"Space Used",
			"Jail Description")

		t.SetLineStyle(table.StyleBrightCyan)
		t.SetDividers(table.UnicodeRoundedDividers)
		t.SetHeaderStyle(table.StyleBold)
	}

	for _, v := range jailsList {
		jailConfig, err := getJailConfig(v, true)
		if err != nil {
			continue
		}

		ID = ID + 1

		jailStatus := ""
		jailOnline, err := checkJailOnline(jailConfig)
		if err != nil {
			return nil
		}

		if jailOnline {
			jailStatus = jailStatus + "🟢"
		} else {
			jailStatus = jailStatus + "🔴"
		}
		if jailConfig.Production {
			jailStatus = jailStatus + "🔁"
		}

		jailRelease, err := getJailReleaseInfo(jailConfig)
		if err != nil {
			return err
		}

		jailUptime, err := getJailUptime(v)
		if err != nil {
			return err
		}

		t.AddRow(strconv.Itoa(ID),
			v,
			jailStatus,
			strconv.Itoa(jailConfig.CPULimitPercent)+"%",
			jailConfig.RAMLimit,
			jailConfig.IPAddress,
			jailRelease,
			jailUptime,
			"Space Used TBD",
			jailConfig.Description,
		)
	}

	t.Render()
	return nil
}
