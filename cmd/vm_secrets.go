package cmd

import (
	HosterVmUtils "HosterCore/internal/pkg/hoster/vm/utils"
	"errors"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"

	"github.com/aquasecurity/table"
	"github.com/spf13/cobra"
)

const VM_DOESNT_EXIST_STRING = "vm doesn't exist on this system"

var (
	vmSecretsUnixTable bool

	vmSecretsCmd = &cobra.Command{
		Use:   "secrets [vmName]",
		Short: "Print out the VM secrets",
		Long:  `Print out the VM secrets, including gwitsuper and root passwords and VNC port+password pairs.`,
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			err := vmSecretsTableOutput(args[0])
			if err != nil {
				log.Fatal(err.Error())
			}
		},
	}
)

func vmSecretsTableOutput(vmName string) error {
	vmFound := false
	vmConf := HosterVmUtils.VmApi{}

	vms, err := HosterVmUtils.ListJsonApi()
	if err != nil {
		return err
	}

	for _, v := range vms {
		if v.Name == vmName {
			vmFound = true
			vmConf = v
		}
	}
	if !vmFound {
		return errors.New(VM_DOESNT_EXIST_STRING)
	}

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
		t.SetHeaders("VM secrets for: " + vmName)
		t.SetHeaderColSpans(0, 3)

		t.AddHeaders(
			"#",
			"Secret Type",
			"Secret Info")

		t.SetLineStyle(table.StyleBrightCyan)
		t.SetDividers(table.UnicodeRoundedDividers)
		t.SetHeaderStyle(table.StyleBold)
	}

	userSecrets, err := readCiUserSecrets(vmConf.Simple.Mountpoint + "/" + vmName)
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

	t.AddRow("1", "VNC Access", fmt.Sprintf("VNC Port: %d || VNC Password: %s", vmConf.VncPort, vmConf.VncPassword))
	t.AddRow("2", "root/administrator password", rootPassword)
	t.AddRow("3", "gwitsuper password", gwitSuperPassword)
	t.Render()

	return nil
}

type UserSecrets struct {
	username string
	password string
}

func readCiUserSecrets(vmFolder string) (r []UserSecrets, e error) {
	vmFolder = strings.TrimSuffix(vmFolder, "/")

	ciUserDataFilePath := vmFolder + "/cloud-init-files/user-data"
	dat, err := os.ReadFile(ciUserDataFilePath)
	if err != nil {
		e = err
		return
	}

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
			r = append(r, userSecret)
		}
		if reGwitsuperPasswordMatch.MatchString(v) {
			userSecret := UserSecrets{}
			userSecret.username = "gwitsuper"
			userSecret.password = reTrim.ReplaceAllString(v, "")
			r = append(r, userSecret)
		}
	}

	return
}
