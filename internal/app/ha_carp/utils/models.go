package CarpUtils

type CarpInfo struct {
	BasePayload
	Status    string `json:"status"`    // Current CARP status: MASTER, BACKUP, INIT
	Interface string `json:"interface"` // Interface name
	Vhid      int    `json:"vhid"`      // Virtual Host Group ID
	Advbase   int    `json:"advbase"`   // Advertisement base interval, seconds
	Advskew   int    `json:"advskew"`   // Advertisement skew, calculated as 1/256th of a second
}

type CarpConfig struct {
	BasePayload
	Interface         string `json:"interface"`           // Interface name
	MasterIpAddress   string `json:"master_ip_address"`   // IP address
	Netmask           string `json:"netmask"`             // Netmask
	Password          string `json:"password"`            // CARP Group Password, used to authenticate CARP packets and prevent unauthorized nodes from joining the CARP group
	Vhid              int    `json:"vhid"`                // Virtual Host Group ID
	Advbase           int    `json:"advbase"`             // Advertisement base interval, seconds
	Advskew           int    `json:"advskew"`             // Advertisement skew, calculated as 1/256th of a second
	ApplyCarpSettings bool   `json:"apply_carp_settings"` // Apply CARP settings to the interface
	FailoverAfter     int    `json:"failover_after"`      // Failover after x seconds
}

type BackupInfo struct {
	BasePayload
	ResourceName     string `json:"resource_name"`     // Resource name
	ResourceType     string `json:"resource_type"`     // Resource type, e.g. "vm", "jail"
	LastSnapshot     string `json:"last_snapshot"`     // Last snapshot name
	ParentHost       string `json:"parent_host"`       // Parent host name
	FailoverStrategy string `json:"failover_strategy"` // Failover strategy, e.g. "cireset" or "change_parent"
}

type HostInfo struct {
	BasePayload
	HostName  string `json:"host_name"`  // Host name
	IpAddress string `json:"ip_address"` // IP address
	LastSeen  int64  `json:"last_seen"`  // Last seen timestamp
	BasicAuth string `json:"basic_auth"` // Basic Auth username:password
	HttpProto string `json:"http_proto"` // HTTP protocol, e.g. "http" or "https"
	HttpPort  int    `json:"http_port"`  // HTTP port
}

type SocketResponse struct {
	BasePayload
	Success bool `json:"success"`
}

type BasePayload struct {
	Type string `json:"type,omitempty"`
}
