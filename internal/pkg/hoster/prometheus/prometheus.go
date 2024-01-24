package HosterPrometheus

import (
	FreeBSDsysctls "HosterCore/internal/pkg/freebsd/sysctls"
	HosterJailUtils "HosterCore/internal/pkg/hoster/jail/utils"
)

// type PrometheusTarget struct {
// 	Targets []string `json:"targets"`
// 	Labels  struct {
// 		Labelname string `json:"labelname"`
// 	} `json:"labels"`
// }

type PrometheusTarget struct {
	Targets []string                 `json:"targets"`
	Labels  []map[string]interface{} `json:"labels"`
}

func GenerateTargets() (r []PrometheusTarget, e error) {
	jails, err := HosterJailUtils.ListAllSimple()
	if err != nil {
		e = err
		return
	}

	hostname, _ := FreeBSDsysctls.SysctlKernHostname()

	for _, v := range jails {
		pt := PrometheusTarget{}
		pt.Targets = append(pt.Targets, v.JailName)

		pt.Labels = append(pt.Labels, map[string]interface{}{"hoster_parent": hostname})
		pt.Labels = append(pt.Labels, map[string]interface{}{"hoster_resource_type": "jail"})
		pt.Labels = append(pt.Labels, map[string]interface{}{"hoster_resource_name": v.JailName})
		pt.Labels = append(pt.Labels, map[string]interface{}{"jail_name": v.JailName})
		pt.Labels = append(pt.Labels, map[string]interface{}{"hoster_resource_encrypted": v.Encrypted})

		r = append(r, pt)
	}

	return
}
