package HosterVmUtils

import "strings"

// Check hardcoded values from OS comments
// Don't forget to update this list when adding new OSes (from the GenerateOsComment function)

func IsOsWindows(input string) bool {
	input = strings.ToLower(input)

	if input == "windows10" || input == "win10" {
		return true
	}
	if input == "windows11" || input == "win11" {
		return true
	}
	if input == "windows-srv19" || input == "winsrv19" || input == "windowssrv19" {
		return true
	}
	if input == "windows-srv22" || input == "winsrv22" || input == "windowssrv22" {
		return true
	}

	// Check if the input contains "windows"
	if strings.Contains(input, "windows") {
		return true
	}

	return false
}

func IsOsFreebsd(input string) bool {
	input = strings.ToLower(input)

	// Check if the input contains "freebsd"
	if strings.Contains(input, "freebsd") {
		return true
	}

	return false
}

func IsOsLinux(input string) bool {
	input = strings.ToLower(input)

	// Check if the input contains "linux"
	if strings.Contains(input, "linux") {
		return true
	}
	// Check if the input contains "debian"
	if strings.Contains(input, "debian") {
		return true
	}
	// Check if the input contains "ubuntu"
	if strings.Contains(input, "ubuntu") {
		return true
	}
	// Check if the input contains "alpine"
	if strings.Contains(input, "alpine") {
		return true
	}
	// Check if the input contains "centos"
	if strings.Contains(input, "centos") {
		return true
	}
	// Check if the input contains "fedora"
	if strings.Contains(input, "fedora") {
		return true
	}
	// Check if the input contains "rhel"
	if strings.Contains(input, "rhel") {
		return true
	}
	// Check if the input contains "rocky"
	if strings.Contains(input, "rocky") {
		return true
	}
	// Check if the input contains "alma"
	if strings.Contains(input, "alma") {
		return true
	}
	// Check if the input contains "suse"
	if strings.Contains(input, "suse") {
		return true
	}

	return false
}
