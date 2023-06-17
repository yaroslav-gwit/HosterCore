package cmd

import (
	"log"
	"os"
	"regexp"
	"strings"

	"github.com/aquasecurity/table"
	"github.com/spf13/cobra"
)

var (
	vmSecretsUnixTable bool

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

	if vmSecretsUnixTable {
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

	userSecrets, err := readCiUserSecrets(vmName)
	if err != nil {
		return err
	}

	rootPassword := ""
	gwitSuperPassword := ""
	for _, v := range userSecrets {
		if v.username == "root" {
			rootPassword = v.password
		}
		if v.username == "gwitsuper" {
			gwitSuperPassword = v.password
		}
	}

	t.AddRow("1", "VNC Access", "VNC Port: "+vmConfigVar.VncPort+" || VNC Password: "+vmConfigVar.VncPassword)
	t.AddRow("2", "root/administrator password", rootPassword)
	t.AddRow("3", "gwitsuper password", gwitSuperPassword)
	t.Render()

	return nil
}

type UserSecrets struct {
	username string
	password string
}

func readCiUserSecrets(vmName string) ([]UserSecrets, error) {
	ciUserDataFilePath := getVmFolder(vmName) + "/cloud-init-files/user-data"
	dat, err := os.ReadFile(ciUserDataFilePath)
	if err != nil {
		return []UserSecrets{}, err
	}

	userSecrets := []UserSecrets{}
	reRootPasswordMatch := regexp.MustCompile(`^root:.*`)
	reGwitsuperPasswordMatch := regexp.MustCompile(`^gwitsuper:.*`)
	reTrim := regexp.MustCompile(`root:|gwitsuper:`)
	for _, v := range strings.Split(string(dat), "\n") {
		if len(v) < 1 {
			continue
		}
		v = strings.TrimSpace(v)
		if reRootPasswordMatch.MatchString(v) {
			userSecret := UserSecrets{}
			userSecret.username = "root"
			userSecret.password = reTrim.ReplaceAllString(v, "")
			userSecrets = append(userSecrets, userSecret)
		}
		if reGwitsuperPasswordMatch.MatchString(v) {
			userSecret := UserSecrets{}
			userSecret.username = "gwitsuper"
			userSecret.password = reTrim.ReplaceAllString(v, "")
			userSecrets = append(userSecrets, userSecret)
		}
	}

	return userSecrets, nil
}
