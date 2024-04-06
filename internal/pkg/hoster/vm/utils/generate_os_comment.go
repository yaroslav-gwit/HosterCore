// Copyright 2024 Hoster Authors. All rights reserved.
// Use of this source code is governed by an Apache License 2.0
// license that can be found in the LICENSE file.

package HosterVmUtils

// Generate a VNC resolution from a pre-set integer.
func GenerateOsComment(input string) (r string) {
	// default case
	r = "Custom OS"

	// case switch
	if input == "debian10" {
		r = "Debian 10"
		return
	}
	if input == "debian11" {
		r = "Debian 11"
		return
	}
	if input == "debian12" {
		r = "Debian 12"
		return
	}

	if input == "ubuntu2004" {
		r = "Ubuntu 20.04"
		return
	}
	if input == "ubuntu2204" {
		r = "Ubuntu 22.04"
		return
	}
	if input == "ubuntu2404" {
		r = "Ubuntu 24.04"
		return
	}

	if input == "almalinux8" {
		r = "AlmaLinux 8"
		return
	}
	if input == "almalinux9" {
		r = "AlmaLinux 9"
		return
	}

	if input == "rockylinux8" {
		r = "RockyLinux 8"
		return
	}
	if input == "rockylinux9" {
		r = "RockyLinux 9"
		return
	}

	if input == "rhel8" {
		r = "RHEL 8"
		return
	}
	if input == "rhel9" {
		r = "RHEL 9"
		return
	}

	if input == "freebsd13ufs" {
		r = "FreeBSD 13 UFS"
		return
	}
	if input == "freebsd13zfs" {
		r = "FreeBSD 13 ZFS"
		return
	}

	if input == "windows10" || input == "win10" {
		r = "Windows 10"
		return
	}
	if input == "windows11" || input == "win11" {
		r = "Windows 11"
		return
	}

	if input == "windows-srv19" || input == "winsrv19" || input == "windowssrv19" {
		r = "Windows Server 19"
		return
	}
	if input == "windows-srv22" || input == "winsrv22" || input == "windowssrv22" {
		r = "Windows Server 22"
		return
	}

	return
}
