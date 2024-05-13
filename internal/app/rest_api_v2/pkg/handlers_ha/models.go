package HandlersHA

import RestApiConfig "HosterCore/internal/app/rest_api_v2/pkg/config"

type HosterHaNode struct {
	LastPing int64                `json:"last_ping"`
	NodeInfo RestApiConfig.HaNode `json:"node_info"`
}

type ModifyHostsDb struct {
	AddOrUpdate bool
	Remove      bool
	Data        HosterHaNode
}
