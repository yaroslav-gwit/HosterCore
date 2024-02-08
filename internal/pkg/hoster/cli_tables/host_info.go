// Copyright 2023 Hoster Authors. All rights reserved.
// Use of this source code is governed by an Apache License 2.0
// license that can be found in the LICENSE file.

package HosterTables

import (
	HosterHostUtils "HosterCore/internal/pkg/hoster/host/utils"
	"fmt"
	"os"

	"github.com/aquasecurity/table"
)

func GenerateHostInfoTable(unixStyleTable bool) error {
	info, err := HosterHostUtils.GetHostInfo()
	if err != nil {
		return nil
	}

	zpoolsHeathy := "Healthy"
	for _, v := range info.ZpoolList {
		if !v.Healthy {
			zpoolsHeathy = "Unhealthy!"
		}
	}

	t := table.New(os.Stdout)
	t.SetLineStyle(table.StyleBrightCyan)
	t.SetDividers(table.UnicodeRoundedDividers)
	t.SetHeaderStyle(table.StyleBold)

	t.SetAlignment(
		table.AlignLeft,   // Hostname
		table.AlignCenter, // Live VMs
		table.AlignCenter, // vCPU:pCPU
		table.AlignCenter, // System Uptime
		table.AlignCenter, // RAM
		table.AlignCenter, // SWAP
		table.AlignCenter, // ARC Size
		table.AlignCenter, // Zpools Health
	)

	t.SetHeaders("Hoster Overview")
	t.SetHeaderColSpans(0, 8)
	t.AddHeaders(
		"Hostname",
		"Live VMs",
		"vCPU:pCPU Ratio",
		"System Uptime",
		"RAM (Used/Total)",
		"SWAP (Used/Total)",
		"ZFS ARC Size",
		"ZPOOL Health",
	)

	t.AddRow(
		info.Hostname,
		fmt.Sprintf("%d", info.LiveVms),
		info.VCPU2PCU,
		info.SystemUptime,
		info.RamInfo.RamUsedHuman+"/"+info.RamInfo.RamOverallHuman,
		info.SwapInfo.SwapUsedHuman+"/"+info.SwapInfo.SwapOverallHuman,
		info.ArcInfo.ArcUsedHuman,
		zpoolsHeathy,
	)

	t.Render()
	return nil
}
