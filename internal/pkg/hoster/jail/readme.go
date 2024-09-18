package HosterJail

import (
	HosterJailUtils "HosterCore/internal/pkg/hoster/jail/utils"
	"fmt"
	"os"
	"strings"
)

// Return Jail's readme markdown file, or an error if something went wrong.
func GetReadme(jailName string) (r string, e error) {
	// If the logger was already set, ignore this
	if !log.ConfigSet {
		log.SetFileLocation(HosterJailUtils.JAIL_AUDIT_LOG_LOCATION)
	}

	jails, err := HosterJailUtils.ListAllSimple()
	if err != nil {
		e = err
		return
	}

	for _, v := range jails {
		if v.JailName == jailName {
			jailFolder := v.Mountpoint + "/" + jailName

			jailReadme := ""
			files, err := os.ReadDir(jailFolder)
			if err != nil {
				e = err
				return
			}

			for _, vv := range files {
				if strings.ToLower(vv.Name()) == "readme.md" {
					jailReadme = jailFolder + "/" + vv.Name()

					file, err := os.ReadFile(jailReadme)
					if err != nil {
						e = err
						return
					}

					r = string(file)
					return
				}
			}

			if len(jailReadme) < 1 {
				e = fmt.Errorf("readme.md not found")
				return
			}
		}
	}

	e = fmt.Errorf("jail was not found")
	return
}
