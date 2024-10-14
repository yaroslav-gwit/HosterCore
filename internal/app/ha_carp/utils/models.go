package CarpUtils

type HaStatus struct {
	BasePayload                // type: ha_status
	Status        string       `json:"status"`         // Current HA status: MASTER, BACKUP, INIT
	CurrentMaster string       `json:"current_master"` // Current master hostname
	Hosts         []HostInfo   `json:"hosts"`          // List of hosts
	Resources     []BackupInfo `json:"resources"`      // List of resources
}

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
	ParticipateInCarp     bool   `json:"participate_in_carp"`     // Apply CARP settings to the interface
	ParticipateInFailover bool   `json:"participate_in_failover"` // Participate in failover
	Vhid                  int    `json:"vhid"`                    // Virtual Host Group ID
	Advbase               int    `json:"advbase"`                 // Advertisement base interval, seconds
	Advskew               int    `json:"advskew"`                 // Advertisement skew, calculated as 1/256th of a second
	FailoverAfter         int    `json:"failover_after"`          // Failover after x seconds
	Interface             string `json:"interface"`               // Interface name
	MasterIpAddress       string `json:"master_ip_address"`       // IP address
	Netmask               string `json:"netmask"`                 // Netmask
	Password              string `json:"password"`                // CARP Group Password, used to authenticate CARP packets and prevent unauthorized nodes from joining the CARP group
}

type BackupInfo struct {
	BasePayload
	ResourceName     string `json:"resource_name"`     // Resource name
	ResourceType     string `json:"resource_type"`     // Resource type, e.g. "vm", "jail"
	LastSnapshot     string `json:"last_snapshot"`     // Last snapshot name
	CurrentHost      string `json:"current_host"`      // Current host name
	ParentHost       string `json:"parent_host"`       // Parent host name
	FailoverStrategy string `json:"failover_strategy"` // Failover strategy, e.g. "cireset" or "change_parent"
}

type HostInfo struct {
	BasePayload
	HostName  string `json:"host_name,omitempty"`  // Host name
	IpAddress string `json:"ip_address,omitempty"` // IP address
	LastSeen  int64  `json:"last_seen,omitempty"`  // Last seen timestamp
}

type SocketResponse struct {
	BasePayload
	Success bool `json:"success"`
}

type BasePayload struct {
	Type string `json:"type,omitempty"`
}

type CarpPingResponse struct {
	Message  string `json:"message"`  // success
	Hostname string `json:"hostname"` // hostname
}
