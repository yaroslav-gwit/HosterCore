package ApiV2client

import (
	CarpUtils "HosterCore/internal/app/ha_carp/utils"
	RestApiConfig "HosterCore/internal/app/rest_api_v2/pkg/config"
	FreeBSDsysctls "HosterCore/internal/pkg/freebsd/sysctls"
	"encoding/json"
	"fmt"
)

// func PingMaster(host CarpUtils.HostInfo) error {
func PingMaster() error {
	carpConfig, err := CarpUtils.ParseCarpConfigFile()
	if err != nil {
		return err
	}

	hostname, err := FreeBSDsysctls.SysctlKernHostname()
	if err != nil {
		return err
	}

	apiConfig, err := RestApiConfig.GetApiConfig()
	if err != nil {
		return err
	}

	url := apiConfig.Protocol + "://" + carpConfig.MasterIpAddress + ":" + fmt.Sprintf("%d", apiConfig.Port) + "/api/v2/carp-ha/ping"
	auth := ""
	for _, v := range apiConfig.HTTPAuth {
		if v.HaUser {
			auth = v.User + ":" + v.Password
		}
	}
	if len(auth) < 1 {
		return fmt.Errorf("no HA user found in the config")
	}

	host := CarpUtils.HostInfo{}
	host.BasicAuth = auth
	host.HostName = hostname
	host.HttpPort = apiConfig.Port
	host.HttpProto = apiConfig.Protocol

	jp, err := json.Marshal(host)
	if err != nil {
		return err
	}
	mp := make(map[string]interface{})
	json.Unmarshal(jp, &mp)

	err = PostFunc(url, auth, mp)
	if err != nil {
		return err
	}

	return nil
}
