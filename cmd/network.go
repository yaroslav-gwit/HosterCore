//go:build freebsd
// +build freebsd

package cmd

import (
	"HosterCore/internal/pkg/emojlog"
	HosterNetwork "HosterCore/internal/pkg/hoster/network"
	"fmt"
	"os"
	"strconv"

	"github.com/aquasecurity/table"
	"github.com/spf13/cobra"
)

var (
	networkCmd = &cobra.Command{
		Use:   "network",
		Short: "Network related operations",
		Long:  `Network related operations.`,
		Run: func(cmd *cobra.Command, args []string) {
			checkInitFile()
			cmd.Help()
		},
	}
)

var (
	networkListUnixStyleTable bool

	networkListCmd = &cobra.Command{
		Use:   "list",
		Short: "List all Hoster related networks",
		Long:  `List all Hoster related networks in a table format.`,
		Run: func(cmd *cobra.Command, args []string) {
			checkInitFile()
			printNetworkInfoTable()
		},
	}
)

var (
	networkInitCmd = &cobra.Command{
		Use:   "init",
		Short: "Initialize or re-load Hoster network configuration",
		Long:  `Initialize or re-load Hoster network configuration.`,
		Run: func(cmd *cobra.Command, args []string) {
			checkInitFile()

			err := loadNetworkConfig()
			if err != nil {
				emojlog.PrintLogMessage("Could not load network config: "+err.Error(), emojlog.Error)
				os.Exit(1)
			}
		},
	}
)

func printNetworkInfoTable() {
	netInfo, err := HosterNetwork.GetNetworkConfig()
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

	if networkListUnixStyleTable {
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

		bridgeInterface := v.BridgeInterface
		if v.BridgeInterface == "None" {
			bridgeInterface = "NAT (no bridge)"
		}

		t.AddRow(
			strconv.Itoa(ID),
			v.NetworkName,
			v.Gateway,
			v.Subnet,
			v.RangeStart+" - "+v.RangeEnd,
			bridgeInterface,
			v.Comment,
		)
	}

	t.Render()
}
