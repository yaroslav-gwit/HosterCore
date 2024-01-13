package HosterJail

import (
	FileExists "HosterCore/pkg/file_exists"
	"strings"
)

// Function takes in a folder path and checks if the jail configuration file exists in it. If it does, it will return true.
//
// For example: JailExists(/zroot/vm-encrypted/jail-test-1) or JailExists(/zroot/vm-encrypted/jail-test-1/)
//
// Trailing "/" is automatically removed.
func JailExists(folderPath string) (r bool) {
	folderPath = strings.TrimSuffix(folderPath, "/")

	if FileExists.CheckUsingOsStat(folderPath + "/" + JAIL_CONFIG_NAME) {
		r = true
		return
	}

	return
}
