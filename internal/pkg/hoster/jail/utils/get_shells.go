package HosterJailUtils

import (
	"os"
	"strings"
)

func GetJailShells(jailName string) (r []string, e error) {
	jail, err := InfoJsonApi(jailName)
	if err != nil {
		e = err
		return
	}

	jailRoot := jail.Simple.Mountpoint + "/" + jailName + "/" + JAIL_ROOT_FOLDER
	etcShells := jailRoot + "/etc/shells"

	file, err := os.ReadFile(etcShells)
	if err != nil {
		e = err
		return
	}

	split := strings.Split(string(file), "\n")
	for _, v := range split {
		v = strings.TrimSpace(v)
		// skip comments
		if strings.HasPrefix(v, "#") {
			continue
		}
		// skip empty lines
		if len(v) < 1 {
			continue
		}
		// append to the result if everything is fine
		r = append(r, v)
	}

	return
}
