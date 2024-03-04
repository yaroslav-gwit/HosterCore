package main

import (
	HosterVmUtils "HosterCore/internal/pkg/hoster/vm/utils"
	"strconv"
)

type VmNumbers struct {
	all               int
	online            int
	backup            int
	offlineProduction int
}

func getVmNumbers() string {
	vmNumbers := VmNumbers{}
	vms, err := HosterVmUtils.ListJsonApi()
	if err != nil {
		return ""
	}

	for _, v := range vms {
		vmNumbers.all += 1
		if v.Running {
			vmNumbers.online += 1
		}
		if v.Backup {
			vmNumbers.backup += 1

			// stop the count here (for this iteration) if the VM is a backup, so we don't receive a false-positive production VM is offline result
			continue
		}
		if v.Production && !v.Running {
			vmNumbers.offlineProduction += 1
		}
	}

	result := "# HELP HosterCore related FreeBSD metrics.\n"
	result = result + "# TYPE hoster gauge\n"
	result = result + "hoster{counter=\"vms_all\"} " + strconv.Itoa(vmNumbers.all) + "\n"
	result = result + "hoster{counter=\"vms_online\"} " + strconv.Itoa(vmNumbers.online) + "\n"
	result = result + "hoster{counter=\"vms_backup\"} " + strconv.Itoa(vmNumbers.backup) + "\n"
	result = result + "hoster{counter=\"vms_offline_in_production\"} " + strconv.Itoa(vmNumbers.offlineProduction) + "\n"

	return result
}
