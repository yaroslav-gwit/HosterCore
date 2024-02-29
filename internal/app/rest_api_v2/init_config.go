package main

import RestApiConfig "HosterCore/internal/app/rest_api_v2/pkg/config"

var restConf RestApiConfig.RestApiConfig

func init() {
	var err error
	restConf, err = RestApiConfig.GetApiConfig()
	if err != nil {
		logInternal.Panicf("could not read the API config: %s", err.Error())
	}
}
