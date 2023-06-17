package cmd

import (
	"log"
	"os"

	"github.com/aquasecurity/table"
	"github.com/spf13/cobra"
)

var (
	vmSecretsCmd = &cobra.Command{
		Use:   "secrets [vmName]",
		Short: "Print out the VM secrets",
		Long:  `Print out the VM secrets, including gwitsuper and root passwords and VNC port+password pairs.`,
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			err := checkInitFile()
			if err != nil {
				log.Fatal(err.Error())
			}

			err = vmSecretsTableOutput(args[0])
			if err != nil {
				log.Fatal(err.Error())
			}
		},
	}
)

func vmSecretsTableOutput(vmName string) error {
	vmConfigVar := vmConfig(vmName)

	var t = table.New(os.Stdout)
	t.SetAlignment(table.AlignRight, //ID
		table.AlignLeft, // Secret Type
		table.AlignLeft) // Secret Info

	if vmSnapshotListUnixStyleTable {
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
		t.SetHeaders("Listing VM secrets for: " + vmName)
		t.SetHeaderColSpans(0, 3)

		t.AddHeaders(
			"ID",
			"Secret Type",
			"Secret Info")

		t.SetLineStyle(table.StyleBrightCyan)
		t.SetDividers(table.UnicodeRoundedDividers)
		t.SetHeaderStyle(table.StyleBold)
	}

	t.AddRow("1", "VNC Access", "VNC Port: "+vmConfigVar.VncPort+" | VNC Password: "+vmConfigVar.VncPassword)
	t.AddRow("2", "root password (administrator if Windows)", "VNC Port: "+vmConfigVar.VncPort+"   VNC Password: "+vmConfigVar.VncPassword)
	t.AddRow("3", "gwitsuper password", "VNC Port: "+vmConfigVar.VncPort+"   VNC Password: "+vmConfigVar.VncPassword)
	t.Render()

	return nil
}
