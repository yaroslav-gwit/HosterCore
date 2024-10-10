package main

import CarpUtils "HosterCore/internal/app/ha_carp/utils"

func init() {
	// Read the configuration file and store it in the global variable
	haConfig, err := CarpUtils.ParseCarpConfigFile()
	if err != nil {
		log.Fatal("Error parsing carp config file:", err)
		return
	}

	activeHaConfig = haConfig
}
