package cmd

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/aquasecurity/table"
	"github.com/spf13/cobra"
)

var (
	hostNetworkCmd = &cobra.Command{
		Use:   "network",
		Short: "Network related operations",
		Long:  `Network related operations.`,
		Run: func(cmd *cobra.Command, args []string) {
			err := checkInitFile()
			if err != nil {
				log.Fatal(err.Error())
			}
			cmd.Help()
		},
	}
)

var (
	hostNetworkInfoUnixStyleTable bool

	hostNetworkInfoCmd = &cobra.Command{
		Use:   "info",
		Short: "Network information output",
		Long:  `Network information output.`,
		Run: func(cmd *cobra.Command, args []string) {
			err := checkInitFile()
			if err != nil {
				log.Fatal(err.Error())
			}
			printNetworkInfoTable()
		},
	}
)

func printNetworkInfoTable() {
	netInfo, err := networkInfo()
	if err != nil {
		fmt.Println(err)
	}

	var ID = 0
	var t = table.New(os.Stdout)
	t.SetAlignment(table.AlignRight, //ID
		table.AlignLeft, // Network Name
		table.AlignLeft, // Gateway
		table.AlignLeft, // Subnet
		table.AlignLeft, // IP Range
		table.AlignLeft, // Bridge Interface
		table.AlignLeft, // Network Comment
	)

	if hostNetworkInfoUnixStyleTable {
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
		t.SetHeaders("Hoster Networks")
		t.SetHeaderColSpans(0, 7)

		t.AddHeaders(
			"#",
			"Network Name",
			"Gateway",
			"Subnet",
			"IP Range",
			"Bridge Interface",
			"Network Comment",
		)

		t.SetLineStyle(table.StyleBrightCyan)
		t.SetDividers(table.UnicodeRoundedDividers)
		t.SetHeaderStyle(table.StyleBold)
	}

	for _, v := range netInfo {
		ID = ID + 1

		bridgeInterface := "NAT (no bridge)"
		if v.ApplyBridgeAddr {
			bridgeInterface = v.BridgeInterface
		}

		t.AddRow(
			strconv.Itoa(ID),
			v.Name,
			v.Gateway,
			v.Subnet,
			v.RangeStart+"-"+v.RangeEnd,
			bridgeInterface,
			v.Comment,
		)
	}

	t.Render()
}
