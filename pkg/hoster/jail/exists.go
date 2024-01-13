package HosterJail

import (
	FileExists "HosterCore/pkg/file_exists"
	"strings"
)

func JailExists(folderPath string) (r bool) {
	folderPath = strings.TrimSuffix(folderPath, "/")

	if FileExists.CheckUsingOsStat(folderPath + "/" + JAIL_CONFIG_NAME) {
		r = true
		return
	}

	return
}
