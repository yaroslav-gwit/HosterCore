// Copyright 2023 Hoster Authors. All rights reserved.
// Use of this source code is governed by an Apache License 2.0
// license that can be found in the LICENSE file.

package HosterLocations

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
		"/usr/local/hoster",
		"/etc/hoster",
		"/root/hoster/config_files",
	}

	return
}
