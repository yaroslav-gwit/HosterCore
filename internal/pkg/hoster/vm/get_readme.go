package HosterVm

import (
	HosterVmUtils "HosterCore/internal/pkg/hoster/vm/utils"
	"fmt"
	"os"
	"strings"
)

// Return VM's readme markdown file, or an error if something went wrong.
func GetReadme(vmName string) (r string, e error) {
	// If the logger was already set, ignore this
	if !log.ConfigSet {
		log.SetFileLocation(HosterVmUtils.VM_AUDIT_LOG_LOCATION)
	}

	vms, err := HosterVmUtils.ListAllSimple()
	if err != nil {
		e = err
		return
	}

	for _, v := range vms {
		if v.VmName == vmName {
			vmFolder := v.Mountpoint + "/" + vmName

			vmReadme := ""
			files, err := os.ReadDir(vmFolder)
			if err != nil {
				e = err
				return
			}

			for _, vv := range files {
				if strings.ToLower(vv.Name()) == "readme.md" {
					vmReadme = vmFolder + "/" + vv.Name()

					file, err := os.ReadFile(vmReadme)
					if err != nil {
						e = err
						return
					}

					r = string(file)
					return
				}
			}

			if len(vmReadme) < 1 {
				e = fmt.Errorf("readme.md not found")
				return
			}
		}
	}

	e = fmt.Errorf("vm not found")
	return
}
