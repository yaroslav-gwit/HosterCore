package cmd

import (
	"HosterCore/internal/pkg/emojlog"
	HosterJail "HosterCore/internal/pkg/hoster/jail"
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

	// jailsList, err := GetAllJailsList()
	// if err != nil {
	// 	return err
	// }

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

	// for _, v := range jailsList {
	// 	jailConfig, err := GetJailConfig(v, true)
	// 	if err != nil {
	// 		fmt.Println("103 fail: " + err.Error())
	// 		continue
	// 	}

	// 	ID = ID + 1

	// 	jailStatus := ""
	// 	jailOnline, err := checkJailOnline(jailConfig)
	// 	if err != nil {
	// 		fmt.Println("112 fail: " + err.Error())
	// 		continue
	// 		// return nil
	// 	}

	// 	if jailOnline {
	// 		jailStatus = jailStatus + "🟢"
	// 	} else {
	// 		jailStatus = jailStatus + "🔴"
	// 	}

	// 	jailDsInfo, err := jailZfsDatasetInfo(jailConfig.ZfsDatasetPath)
	// 	if err != nil {
	// 		fmt.Println("125 fail: " + err.Error())
	// 		continue
	// 		// return err
	// 	}
	// 	if jailDsInfo.Encrypted {
	// 		jailStatus = jailStatus + "🔒"
	// 	}

	// 	if jailConfig.Production {
	// 		jailStatus = jailStatus + "🔁"
	// 	}

	// 	jailRelease, err := getJailReleaseInfo(jailConfig)
	// 	if err != nil {
	// 		fmt.Println("139 fail: " + err.Error())
	// 		continue
	// 		// return err
	// 	}

	// 	jailUptime := getJailUptime(v)

	// 	t.AddRow(strconv.Itoa(ID),
	// 		v,
	// 		jailStatus,
	// 		strconv.Itoa(jailConfig.CPULimitPercent)+"%",
	// 		jailConfig.RAMLimit,
	// 		jailConfig.IPAddress,
	// 		jailRelease,
	// 		jailUptime,
	// 		jailDsInfo.StorageUsedHuman,
	// 		jailConfig.Description,
	// 	)
	// }
	for _, v := range jailList {
		t.AddRow(strconv.Itoa(ID),
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

// type JailZfsDatasetStruct struct {
// 	Encrypted        bool
// 	StorageUsedHuman string
// 	StorageUsedBytes int
// }

// func jailZfsDatasetInfo(zfsDatasetPath string) (zfsDsInfo JailZfsDatasetStruct, zfsDsError error) {
// 	zfsListOutput, err := exec.Command("zfs", "list", "-Hp", "-o", "name,encryption,used", zfsDatasetPath).CombinedOutput()
// 	//    [0]                               [1]          [2]
// 	// zroot/vm-encrypted/wordpress-one	aes-256-gcm	1244692480
// 	if err != nil {
// 		errorValue := "FATAL: " + string(zfsListOutput) + "; " + err.Error()
// 		zfsDsError = errors.New(errorValue)
// 		return
// 	}

// 	reSpaceSplit := regexp.MustCompile(`\s+`)
// 	for _, v := range strings.Split(string(zfsListOutput), "\n") {
// 		tempSplitList := reSpaceSplit.Split(v, -1)
// 		if len(tempSplitList) <= 1 {
// 			continue
// 		}

// 		if tempSplitList[1] == "off" {
// 			zfsDsInfo.Encrypted = false
// 		} else {
// 			zfsDsInfo.Encrypted = true
// 		}

// 		zfsDsInfo.StorageUsedBytes, err = strconv.Atoi(tempSplitList[2])
// 		if err != nil {
// 			zfsDsError = err
// 			return
// 		}

// 		zfsDsInfo.StorageUsedHuman = ByteConversion(zfsDsInfo.StorageUsedBytes)
// 		return
// 	}

// 	return
// }
