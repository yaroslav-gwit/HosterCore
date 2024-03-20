// Copyright 2023 Hoster Authors. All rights reserved.
// Use of this source code is governed by an Apache License 2.0
// license that can be found in the LICENSE file.

package HosterLocations

import (
	FileExists "HosterCore/internal/pkg/file_exists"
	"fmt"
)

func GetBinaryFolders() (r []string) {
	r = []string{
		"/opt/hoster-core",
		"/opt/hoster",
		"/usr/local/bin",
		"/bin",
		"/root/hoster",
	}

	return
}

func GetConfigFolders() (r []string) {
	r = []string{
		"/opt/hoster-core/config_files",
		"/opt/hoster/config_files",
		"/usr/local/etc/hoster",
		"/etc/hoster",
		"/root/hoster/config_files",
	}

	return
}

func LocateBinary(binaryName string) (r string, e error) {
	for _, v := range GetBinaryFolders() {
		if FileExists.CheckUsingOsStat(v + "/" + binaryName) {
			r = v + "/" + binaryName
			return
		}
	}

	if len(r) < 1 {
		e = fmt.Errorf("could not locate the binary %s", binaryName)
		return
	}

	return
}
