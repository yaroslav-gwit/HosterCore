package CarpUtils

import (
	HosterLocations "HosterCore/internal/pkg/hoster/locations"
	"encoding/json"
	"os"
)

func ParseCarpConfigFile() (r CarpConfig, e error) {
	fileLoc, err := HosterLocations.LocateConfig("carp.json")
	if err != nil {
		e = err
		return
	}

	config, err := os.ReadFile(fileLoc)
	if err != nil {
		e = err
		return
	}

	err = json.Unmarshal(config, &r)
	if err != nil {
		e = err
		return
	}

	return
}
