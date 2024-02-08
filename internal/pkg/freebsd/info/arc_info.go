// Copyright 2023 Hoster Authors. All rights reserved.
// Use of this source code is governed by an Apache License 2.0
// license that can be found in the LICENSE file.

package FreeBSDOsInfo

import (
	"HosterCore/internal/pkg/byteconversion"
	FreeBSDsysctls "HosterCore/internal/pkg/freebsd/sysctls"
)

type ArcInfo struct {
	ArcUsedHuman string `json:"arc_used_human"`
	ArcUsedBytes uint64 `json:"arc_used_bytes"`
}

func GetArcInfo() (r ArcInfo, e error) {
	b, err := FreeBSDsysctls.SysctlKstatZfsMiscArcstatsSize()
	if err != nil {
		e = err
		return
	}

	r.ArcUsedBytes = b
	r.ArcUsedHuman = byteconversion.BytesToHuman(b)

	return
}
