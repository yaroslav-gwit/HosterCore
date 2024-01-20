package HosterHostUtils

import (
	"os/exec"
	"regexp"
	"strings"
)

func GetMajorFreeBsdRelease() (release string, releaseError error) {
	out, err := exec.Command("uname", "-r").CombinedOutput()
	if err != nil {
		releaseError = err
		return
	}

	release = strings.TrimSpace(string(out))

	// Strip minor patch version
	reStripMinor := regexp.MustCompile(`-p.*`)
	release = reStripMinor.ReplaceAllString(release, "")
	// EOF Strip minor patch version

	return
}
