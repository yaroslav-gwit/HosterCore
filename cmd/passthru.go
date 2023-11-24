package cmd

import (
	"HosterCore/emojlog"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/aquasecurity/table"
	"github.com/spf13/cobra"
)

var (
	passthruCmd = &cobra.Command{
		Use:   "passthru",
		Short: "Bhyve passthru related commands",
		Long:  `Bhyve passthru related commands.`,
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			err := checkInitFile()
			if err != nil {
				emojlog.PrintLogMessage(err.Error(), emojlog.Error)
				os.Exit(1)
			}

			cmd.Help()
		},
	}
)

var (
	passthruListCmd = &cobra.Command{
		Use:   "list",
		Short: "List the passthru-ready devices",
		Long:  `List the passthru-ready devices.`,
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			err := checkInitFile()
			if err != nil {
				emojlog.PrintLogMessage(err.Error(), emojlog.Error)
				os.Exit(1)
			}

			printPptDevicesTable()
		},
	}
)

type PPT struct {
	PptDevice      string
	PciIdRaw       string
	PciIdForBhyve  string
	Vendor         string
	Model          string
	Type           string
	StatusBusyBool bool
	StatusString   string
}

func printPptDevicesTable() {
	pptTableData := pptDevices()

	outputTable := table.New(os.Stdout)
	outputTable.SetHeaders("Bhyve Passthru Devices")
	outputTable.SetHeaderColSpans(0, 7)
	outputTable.AddHeaders("PPT Device", "PCI ID Raw", "Type", "Bhyve PCI ID", "Vendor", "Model", "Status")
	outputTable.SetLineStyle(table.StyleBrightCyan)
	outputTable.SetDividers(table.UnicodeRoundedDividers)
	outputTable.SetHeaderStyle(table.StyleBold)
	for _, v := range pptTableData {
		outputTable.AddRow(v.PptDevice, v.PciIdRaw, v.Type, v.PciIdForBhyve, v.Vendor, v.Model, v.StatusString)
	}

	outputTable.Render()
}

func pptDevices() []PPT {
	out, err := exec.Command("pciconf", "-l").CombinedOutput()
	if err != nil {
		log.Fatal(string(out) + err.Error())
	}

	pptDevices := []string{}
	reMatchDriver := regexp.MustCompile(`^ppt`)
	for _, v := range strings.Split(string(out), "\n") {
		v = strings.TrimSpace(v)
		if reMatchDriver.MatchString(v) {
			pptDevices = append(pptDevices, v)
		}
	}

	if len(pptDevices) < 1 {
		log.Fatal("No devices were found!")
	}

	pptDevicesLong := []string{}
	for _, v := range pptDevices {
		v = strings.Split(v, "@")[0]
		v = strings.TrimSpace(v)
		out, err := exec.Command("pciconf", "-lv", v).CombinedOutput()
		if err != nil {
			log.Fatal(string(out) + err.Error())
		}
		pptDevicesLong = append(pptDevicesLong, string(out))
	}

	busyPptDevices := usedPptDevices()

	finalResult := []PPT{}
	reMatchPci := regexp.MustCompile(`pci.*?:.*?:.*?:.*?:`)
	reMatchVendor := regexp.MustCompile(`^vendor`)
	reMatchModel := regexp.MustCompile(`^device`)
	reMatchClass := regexp.MustCompile(`^class`)
	reMatchSubClass := regexp.MustCompile(`^subclass`)
	for _, v := range pptDevicesLong {
		singlePpt := PPT{}
		for _, vv := range strings.Split(v, "\n") {
			if reMatchDriver.MatchString(vv) {
				vvCopy := vv
				vvCopy = strings.Split(vvCopy, "@")[0]
				vvCopy = strings.TrimSpace(vvCopy)
				singlePpt.PptDevice = vvCopy

				singlePpt.PciIdRaw = reMatchPci.FindString(vv)
				pciDev := reMatchPci.FindString(vv)
				pciDev = strings.TrimPrefix(pciDev, "pci0:")
				pciDev = strings.TrimSuffix(pciDev, ":")
				pciDev = strings.ReplaceAll(pciDev, ":", "/")
				singlePpt.PciIdForBhyve = pciDev
				continue
			}

			if reMatchVendor.MatchString(strings.TrimSpace(vv)) {
				vv = strings.Split(vv, "=")[1]
				vv = strings.TrimSpace(vv)
				vv = strings.Trim(vv, "'")
				singlePpt.Vendor = vv
				continue
			}

			reStripBrackets := regexp.MustCompile(`\[|]`)
			if reMatchModel.MatchString(strings.TrimSpace(vv)) {
				vv = strings.Split(vv, "=")[1]
				vv = strings.TrimSpace(vv)
				vv = strings.Trim(vv, "'")
				vv = reStripBrackets.ReplaceAllString(vv, "")
				singlePpt.Model = vv
				continue
			}

			if reMatchClass.MatchString(strings.TrimSpace(vv)) {
				vv = strings.Split(vv, "=")[1]
				vv = strings.TrimSpace(vv)
				singlePpt.Type = vv
			}
			if reMatchSubClass.MatchString(strings.TrimSpace(vv)) {
				vv = strings.Split(vv, "=")[1]
				vv = strings.TrimSpace(vv)
				singlePpt.Type = singlePpt.Type + "/" + vv
				continue
			}
		}

		for _, v := range busyPptDevices {
			if v.Dev == singlePpt.PciIdForBhyve {
				singlePpt.StatusBusyBool = true
				singlePpt.StatusString = "In use by: " + v.VmName
			}
		}
		if !singlePpt.StatusBusyBool {
			singlePpt.StatusString = "Available"
		}

		finalResult = append(finalResult, singlePpt)
	}

	return finalResult
}

type UsedPptDevice struct {
	VmName string
	Dev    string
}

func usedPptDevices() (usedPptDevices []UsedPptDevice) {
	for _, vm := range getAllVms() {
		config := vmConfig(vm)
		for _, pptDev := range config.Passthru {
			usedPptDev := UsedPptDevice{}
			usedPptDev.Dev = pptDev
			usedPptDev.VmName = vm
			usedPptDevices = append(usedPptDevices, usedPptDev)
		}
	}

	return
}
