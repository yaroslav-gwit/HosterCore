package ApiV2client

import (
	CarpUtils "HosterCore/internal/app/ha_carp/utils"
	RestApiConfig "HosterCore/internal/app/rest_api_v2/pkg/config"
	FreeBSDsysctls "HosterCore/internal/pkg/freebsd/sysctls"
	"encoding/json"
	"fmt"
)

func PingMaster(carpConfig CarpUtils.CarpConfig) (r string, e error) {
	hostname, err := FreeBSDsysctls.SysctlKernHostname()
	if err != nil {
		e = err
		return
	}

	apiConfig, err := RestApiConfig.GetApiConfig()
	if err != nil {
		e = err
		return
	}

	url := apiConfig.Protocol + "://" + carpConfig.MasterIpAddress + ":" + fmt.Sprintf("%d", apiConfig.Port) + "/api/v2/carp-ha/ping"
	auth := ""
	for _, v := range apiConfig.HTTPAuth {
		if v.HaUser {
			auth = v.User + ":" + v.Password
		}
	}
	if len(auth) < 1 {
		e = fmt.Errorf("no HA user found in the config")
		return
	}

	host := CarpUtils.HostInfo{}
	host.HostName = hostname
	jp, err := json.Marshal(host)
	if err != nil {
		e = err
		return
	}
	mp := make(map[string]interface{})
	json.Unmarshal(jp, &mp)

	body, err := PostFunc(url, auth, mp)
	if err != nil {
		e = err
		return
	}

	res := CarpUtils.CarpPingResponse{}
	err = json.Unmarshal(body, &res)
	if err != nil {
		e = err
		return
	}

	r = res.Hostname
	return
}

func SendLocalState(haState CarpUtils.HaStatus, remoteIp string, masterHostname string) error {
	apiConfig, err := RestApiConfig.GetApiConfig()
	if err != nil {
		return err
	}

	url := apiConfig.Protocol + "://" + remoteIp + ":" + fmt.Sprintf("%d", apiConfig.Port) + "/api/v2/carp-ha/receive-state/" + masterHostname
	auth := ""
	for _, v := range apiConfig.HTTPAuth {
		if v.HaUser {
			auth = v.User + ":" + v.Password
		}
	}
	if len(auth) < 1 {
		return fmt.Errorf("no HA user found in the config")
	}

	jp, err := json.Marshal(haState)
	if err != nil {
		return err
	}
	mp := make(map[string]interface{})
	json.Unmarshal(jp, &mp)

	_, err = PostFunc(url, auth, mp)
	if err != nil {
		return err
	}

	return nil
}
