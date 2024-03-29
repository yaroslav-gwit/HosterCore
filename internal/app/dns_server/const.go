// Copyright 2024 Hoster Authors. All rights reserved.
// Use of this source code is governed by an Apache License 2.0
// license that can be found in the LICENSE file.

package main

// Hardcoded failover DNS servers (in case user's main DNS server fails)
const DNS_SRV4_QUAD_NINE = "9.9.9.9:53"
const DNS_SRV4_CLOUD_FLARE = "1.1.1.1:53"

const (
	LOG_SUPERVISOR = "supervisor"
	LOG_SYS_OUT    = "sys_stdout"
	LOG_SYS_ERR    = "sys_stderr"
	LOG_DNS_LOCAL  = "dns_locals"
	LOG_DNS_GLOBAL = "dns_global"
	LOG_DEV_DEBUG  = "dev_debug"
)
