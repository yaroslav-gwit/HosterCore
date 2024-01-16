// Copyright 2023s The Hoster Authors. All rights reserved.
// Use of this source code is governed by a Apache License 2.0
// license that can be found in the LICENSE fil

package hostconfig

type (
	SSHKey struct {
		KeyValue string `json:"key_value"`
		Comment  string `json:"comment"`
	}

	ConsolePanel struct {
		PIN                string `json:"pin"`
		MaxPINLength       int    `json:"max_pin_length"`
		MaximumPINAttempts int    `json:"maximum_pin_attempts"`
		LockTimeout        int    `json:"lock_timeout"`
		SessionTime        int    `json:"session_time"`
	}

	Config struct {
		ImageServer    string       `json:"public_vm_image_server"`
		BackupServers  []string     `json:"backup_servers"`
		ActiveDatasets []string     `json:"active_datasets"`
		DnsServers     []string     `json:"dns_servers,omitempty"`   // this is a new field, that might not be implemented on all of the nodes yet
		HostDNSACLs    []string     `json:"host_dns_acls,omitempty"` // this field is deprecated, remains here for backwards compatibility, and will be removed at some point
		HostSSHKeys    []SSHKey     `json:"host_ssh_keys"`
		ConsolePanel   ConsolePanel `json:"console_panel"`
	}
)
