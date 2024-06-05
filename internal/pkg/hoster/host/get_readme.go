package HosterHost

import (
	HosterLocations "HosterCore/internal/pkg/hoster/locations"
	"os"
)

// Returns the content of the readme.md file for this particular Hoster node.
func GetReadme() (r string, e error) {
	loc, err := HosterLocations.LocateConfigCaseInsensitive("readme.md")
	if err != nil {
		e = err
		return
	}

	file, err := os.ReadFile(loc)
	if err != nil {
		e = err
		return
	}

	r = string(file)
	return
}
