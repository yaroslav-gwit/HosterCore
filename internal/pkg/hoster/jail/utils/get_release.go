package HosterJailUtils

import (
	FileExists "HosterCore/internal/pkg/file_exists"
	"os"
	"regexp"
	"strings"
)

// This function takes in a Jail path, like this one: /tank/vm-encrypted/prometheus/,
// appends the Jail root folder to it, so it becomes /tank/vm-encrypted/prometheus/root_folder.
//
// Then checks `/etc/os-release` inside of the jail, and parses out the FreeBSD release from it (eg 13.2-RELEASE).
//
// Returns "-", if the Jail was just deployed, because the `/etc/os-release` is missing.
func ReleaseVersion(jailFolder string) (r string, e error) {
	jailFolder = strings.TrimSuffix(jailFolder, "/")
	var jailOsReleaseFile []byte
	var err error

	jailRootPath := jailFolder + "/" + JAIL_ROOT_FOLDER
	if FileExists.CheckUsingOsStat(jailRootPath + "/etc/os-release") {
		jailOsReleaseFile, err = os.ReadFile(jailRootPath + "/etc/os-release")
		if err != nil {
			e = err
			return
		}
	} else {
		r = "-"
	}

	reMatchVersion := regexp.MustCompile(`VERSION=`)
	reMatchQuotes := regexp.MustCompile(`"`)

	for _, v := range strings.Split(string(jailOsReleaseFile), "\n") {
		if reMatchVersion.MatchString(v) {
			v = reMatchVersion.ReplaceAllString(v, "")
			v = reMatchQuotes.ReplaceAllString(v, "")
			r = v
			return
		}
	}

	return
}
