// Copyright 2024 Hoster Authors. All rights reserved.
// Use of this source code is governed by an Apache License 2.0
// license that can be found in the LICENSE file.

package HosterVmUtils

type ListTable struct {
	VmName        string
	VmStatus      string
	CPUSockets    int
	CPUCores      int
	VmMemory      string
	MainIpAddress string
	OsType        string
	OsComment     string
	VmUptime      string
	DiskUsedTotal string
	VmDescription string
}

func ListAllTable() (r []ListTable, e error) {
	vms, err := ListJsonApi()
	if err != nil {
		e = err
		return
	}

	for _, v := range vms {
		l := ListTable{}
		l.VmName = v.Name

		if v.Running {
			l.VmStatus = l.VmStatus + "🟢"
		}
		if !v.Running && !v.Backup {
			l.VmStatus = l.VmStatus + "🔴"
		}
		if v.Backup {
			l.VmStatus = l.VmStatus + "💾"
		}
		if v.Encrypted {
			l.VmStatus = l.VmStatus + "🔒"
		}
		if v.Production {
			l.VmStatus = l.VmStatus + "🔁"
		}

		l.CPUCores = v.CPUCores
		l.CPUSockets = v.CPUSockets
		l.VmMemory = v.Memory
		l.OsType = v.OsType
		l.OsComment = v.OsComment
		l.VmUptime = v.Uptime
		l.VmDescription = v.Description

		if len(v.Networks) > 0 {
			l.MainIpAddress = v.Networks[0].IPAddress
		} else {
			l.MainIpAddress = "N/A"
		}

		if len(v.Disks) > 0 {
			l.DiskUsedTotal = v.Disks[0].UsedHuman + "/" + v.Disks[0].TotalHuman
		} else {
			l.DiskUsedTotal = "N/A"
		}

		r = append(r, l)
	}

	return
}
