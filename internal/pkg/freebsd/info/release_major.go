package FreeBSDOsInfo

import (
	"fmt"
	"os/exec"
	"regexp"
	"strings"
)

// Returns a major FreeBSD release version,
// which is essentially a stripped down version of `uname -r` minus the `p-*` part.
//
// For example, a major release looks like this: 13.2-RELEASE, NOT like this: 13.2-RELEASE-p9
func GetMajorReleaseVersion() (r string, e error) {
	out, err := exec.Command("uname", "-r").CombinedOutput()
	if err != nil {
		e = fmt.Errorf("error: %s; %s", strings.TrimSpace(string(out)), err.Error())
		return
	}
	r = strings.TrimSpace(string(out))

	// Strip minor patch version
	reStripMinor := regexp.MustCompile(`-p.*`)
	r = reStripMinor.ReplaceAllString(r, "")
	// EOF Strip minor patch version

	return
}
