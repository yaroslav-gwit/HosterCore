package HosterPrometheus

import (
	FreeBSDsysctls "HosterCore/internal/pkg/freebsd/sysctls"
	HosterVmUtils "HosterCore/internal/pkg/hoster/vm/utils"
	"strings"
)

type PrometheusLabels struct {
	HosterParent            string `json:"hoster_parent"`
	HosterResourceType      string `json:"hoster_resource_type"`
	HosterResourceName      string `json:"hoster_resource_name"`
	HosterVmName            string `json:"hoster_vm_name,omitempty"`
	HosterJailName          string `json:"hoster_jail_name,omitempty"`
	HosterResourceEncrypted string `json:"hoster_resource_encrypted"`
}

type PrometheusTarget struct {
	Targets []string         `json:"targets"`
	Labels  PrometheusLabels `json:"labels"`
	// Labels  []map[string]string `json:"labels"`
}

// This function generates a list of Prometheus targets using the VM DNS names.
func GenerateVmTargets(useIps bool) (r []PrometheusTarget, e error) {
	hostname, _ := FreeBSDsysctls.SysctlKernHostname()
	// vms, err := HosterVmUtils.ListJsonApi()
	vms, err := HosterVmUtils.ReadCache()
	if err != nil {
		e = err
		return
	}

	for _, v := range vms {
		if v.Backup {
			continue
		}
		if !v.Running {
			continue
		}

		pt := PrometheusTarget{}

		target := generateVmTarget(v, useIps)
		pt.Targets = append(pt.Targets, target)

		pl := PrometheusLabels{HosterParent: hostname, HosterResourceType: "vm", HosterResourceName: v.Name, HosterVmName: v.Name}
		if v.Encrypted {
			pl.HosterResourceEncrypted = "true"
		} else {
			pl.HosterResourceEncrypted = "false"
		}
		// pt.Labels = append(pt.Labels, map[string]interface{}{"hoster_resource_encrypted": v.Encrypted})

		pt.Labels = pl
		r = append(r, pt)
	}

	return
}

func generateVmTarget(info HosterVmUtils.VmApi, useIps bool) string {
	address := ""
	if useIps {
		for _, v := range info.Networks {
			address = v.IPAddress
			break // pick the first available IP address
		}
	}
	if len(address) < 1 {
		address = info.Name
	}

	if strings.Contains(info.OsType, "windows") {
		return address + ":9182"
	}
	if strings.Contains(info.OsType, "winsrv") {
		return address + ":9182"
	}

	return address + ":9100"
}
