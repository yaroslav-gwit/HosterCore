// Copyright 2024 Hoster Authors. All rights reserved.
// Use of this source code is governed by an Apache License 2.0
// license that can be found in the LICENSE file.

package HosterVmUtils

const CiUserDataTemplate = `#cloud-config

users:
  - default
  - name: root
    lock_passwd: false
    ssh_pwauth: true
    disable_root: false
    ssh_authorized_keys:
	  {{- range .SshKeys}}
      - {{ .KeyValue }}
	  {{- end }}

  - name: gwitsuper
    sudo: ALL=(ALL) NOPASSWD:ALL
    groups: wheel
    ssh_pwauth: true
    lock_passwd: false
    ssh_authorized_keys:
	  {{- range .SshKeys}}
      - {{ .KeyValue }}
	  {{- end }}

chpasswd:
  list: |
    root:{{ .RootPassword }}
    gwitsuper:{{ .GwitsuperPassword }}
  expire: False

package_update: false
package_upgrade: false
`

const CiMetaDataTemplate = `instance-id: iid-{{ .InstanceId }}
local-hostname: {{ .VmName }}
`

const CiNetworkConfigTemplate = `version: 2
ethernets:
  interface0:
    match:
      macaddress: "{{ .MacAddress }}"

    {{- if not (or (eq .OsType "freebsd13ufs") (eq .OsType "freebsd13zfs")) }} 
    set-name: eth0
	{{- end }}
    addresses:
    - {{ .IpAddress }}/{{ .NakedSubnet }}
 
    gateway4: {{ .Gateway }}
 
    nameservers:
      search: [ {{ .ParentHost }}.internal.lan, ]
      addresses: [{{ .DnsServer }}, ]
`
