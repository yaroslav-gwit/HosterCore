#!/usr/bin/env bash

#_ CHECK IF USER IS ROOT _#
if [ "$EUID" -ne 0 ]; then echo " 🚦 ERROR: Please run this script as root user!" && exit 1; fi

#_ SET DEFAULT VARS _#
NETWORK_NAME="${DEF_NETWORK_NAME:=internal}"
NETWORK_BR_ADDR="${DEF_NETWORK_BR_ADDR:=10.0.101.254}"
NETWORK_SUBNET="${DEF_NETWORK_SUBNET:=10.0.101.0/24}"
NETWORK_RANGE_START="${DEF_NETWORK_RANGE_START:=10.0.101.10}"
NETWORK_RANGE_END="${DEF_NETWORK_RANGE_END:=10.0.101.200}"
PUBLIC_INTERFACE="${DEF_PUBLIC_INTERFACE:=$(ifconfig | head -1 | awk '{ print $1 }' | sed s/://)}"
UPSTREAM_DNS_SERVER="${DEF_UPSTREAM_DNS_SERVER:=1.1.1.2}"

#_ SET WORKING DIRECTORY _#
HOSTER_WD="/opt/hoster-core/"

#_ INSTALL THE REQUIRED PACKAGES _#
pkg update
pkg upgrade -y
pkg install -y vim bash pftop tmux qemu-tools git openssl curl
pkg install -y bhyve-firmware uefi-edk2-bhyve-csm edk2-bhyve
pkg install -y htop wget gtar unzip cdrkit-genisoimage go beadm

#_ OPTIONAL PACKAGES _#
# (install for easier debugging)
# pkg install -y nano micro bmon iftop mc fusefs-sshfs gnu-watch fping fish bhyve-rc grub2-bhyve

if [[ -f /bin/bash ]]; then rm /bin/bash; fi
ln "$(which bash)" /bin/bash

#_ SET ENCRYPTED ZFS PASSWORD _#
if [ -z "${DEF_ZFS_ENCRYPTION_PASSWORD}" ]; then
    ZFS_RANDOM_PASSWORD=$(openssl rand -base64 32 | tr -dc '[:alnum:]')
else
    ZFS_RANDOM_PASSWORD=${DEF_ZFS_ENCRYPTION_PASSWORD}
fi

#_ GENERATE SSH KEYS _#
if [[ ! -f /root/.ssh/id_rsa ]]; then
    ssh-keygen -b 4096 -t rsa -f /root/.ssh/id_rsa -q -N ""
else
    echo " 🔷 DEBUG: SSH key was found, no need to generate a new one"
fi

if [[ ! -f /root/.ssh/config ]]; then
    touch /root/.ssh/config && chmod 600 /root/.ssh/config
fi

HOST_SSH_KEY=$(cat /root/.ssh/id_rsa.pub)

#_ REGISTER IF REQUIRED DATASETS EXIST _#
ENCRYPTED_DS=$(zfs list | grep -c "zroot/vm-encrypted")
UNENCRYPTED_DS=$(zfs list | grep -c "zroot/vm-unencrypted")

#_ CREATE ZFS DATASETS IF THEY DON'T EXIST _#
if [[ ${ENCRYPTED_DS} -lt 1 ]]; then
    zpool set autoexpand=on zroot
    zpool set autoreplace=on zroot
    zfs set primarycache=metadata zroot
    echo -e "${ZFS_RANDOM_PASSWORD}" | zfs create -o encryption=on -o keyformat=passphrase zroot/vm-encrypted
fi

if [[ ${UNENCRYPTED_DS} -lt 1 ]]; then
    zpool set autoexpand=on zroot
    zpool set autoreplace=on zroot
    zfs set primarycache=metadata zroot
    zfs create zroot/vm-unencrypted
fi

#_ BOOTLOADER OPTIMISATIONS _#
BOOTLOADER_FILE="/boot/loader.conf"
# CMD_LINE='fusefs_load="YES"' && if [[ $(grep -c "${CMD_LINE}" ${BOOTLOADER_FILE}) -lt 1 ]]; then echo "${CMD_LINE}" >> ${BOOTLOADER_FILE}; fi
CMD_LINE='vm.kmem_size="400M"' && if [[ $(grep -c "${CMD_LINE}" ${BOOTLOADER_FILE}) -lt 1 ]]; then echo "${CMD_LINE}" >>${BOOTLOADER_FILE}; fi
CMD_LINE='vm.kmem_size_max="400M"' && if [[ $(grep -c "${CMD_LINE}" ${BOOTLOADER_FILE}) -lt 1 ]]; then echo "${CMD_LINE}" >>${BOOTLOADER_FILE}; fi
CMD_LINE='vfs.zfs.arc_max="40M"' && if [[ $(grep -c "${CMD_LINE}" ${BOOTLOADER_FILE}) -lt 1 ]]; then echo "${CMD_LINE}" >>${BOOTLOADER_FILE}; fi
CMD_LINE='vfs.zfs.vdev.cache.size="5M"' && if [[ $(grep -c "${CMD_LINE}" ${BOOTLOADER_FILE}) -lt 1 ]]; then echo "${CMD_LINE}" >>${BOOTLOADER_FILE}; fi
# CMD_LINE='virtio_blk_load="YES"' && if [[ $(grep -c "${CMD_LINE}" ${BOOTLOADER_FILE}) -lt 1 ]]; then echo "${CMD_LINE}" >> ${BOOTLOADER_FILE}; fi
CMD_LINE='pf_load="YES"' && if [[ $(grep -c "${CMD_LINE}" ${BOOTLOADER_FILE}) -lt 1 ]]; then echo "${CMD_LINE}" >>${BOOTLOADER_FILE}; fi
CMD_LINE='kern.racct.enable=1' && if [[ $(grep -c "${CMD_LINE}" ${BOOTLOADER_FILE}) -lt 1 ]]; then echo "${CMD_LINE}" >>${BOOTLOADER_FILE}; fi

#_ PF CONFIG BLOCK IN rc.conf _#
RC_CONF_FILE="/etc/rc.conf"
CMD_LINE='pf_enable="yes"' && if [[ $(grep -c "${CMD_LINE}" ${RC_CONF_FILE}) -lt 1 ]]; then echo "${CMD_LINE}" >>${RC_CONF_FILE}; fi
CMD_LINE='pf_rules="/etc/pf.conf"' && if [[ $(grep -c "${CMD_LINE}" ${RC_CONF_FILE}) -lt 1 ]]; then echo "${CMD_LINE}" >>${RC_CONF_FILE}; fi
CMD_LINE='pflog_enable="yes"' && if [[ $(grep -c "${CMD_LINE}" ${RC_CONF_FILE}) -lt 1 ]]; then echo "${CMD_LINE}" >>${RC_CONF_FILE}; fi
CMD_LINE='pflog_logfile="/var/log/pflog"' && if [[ $(grep -c "${CMD_LINE}" ${RC_CONF_FILE}) -lt 1 ]]; then echo "${CMD_LINE}" >>${RC_CONF_FILE}; fi
CMD_LINE='pflog_flags=""' && if [[ $(grep -c "${CMD_LINE}" ${RC_CONF_FILE}) -lt 1 ]]; then echo "${CMD_LINE}" >>${RC_CONF_FILE}; fi
CMD_LINE='gateway_enable="yes"' && if [[ $(grep -c "${CMD_LINE}" ${RC_CONF_FILE}) -lt 1 ]]; then echo "${CMD_LINE}" >>${RC_CONF_FILE}; fi
# CMD_LINE='rtclocaltime="NO"' && if [[ $(grep -c "${CMD_LINE}" ${RC_CONF_FILE}) -lt 1 ]]; then echo "${CMD_LINE}" >> ${RC_CONF_FILE}; fi

#_ SET CORRECT PROFILE FILE _#
cat <<'EOF' | cat >/root/.profile
PATH=/sbin:/bin:/usr/sbin:/usr/bin:/usr/local/sbin:/usr/local/bin:~/bin:/opt/hoster-core
export PATH
HOME=/root
export HOME
TERM=${TERM:-xterm}
export TERM
PAGER=less
export PAGER

# set ENV to a file invoked each time sh is started for interactive use.
ENV=$HOME/.shrc; export ENV

# Query terminal size; useful for serial lines.
if [ -x /usr/bin/resizewin ] ; then /usr/bin/resizewin -z ; fi

# Uncomment to display a random cookie on each login.
# if [ -x /usr/bin/fortune ] ; then /usr/bin/fortune -s ; fi

[ -z "$PS1" ] && true || echo "Hoster version: $(/opt/hoster-core/hoster version)"
export EDITOR=vim

EOF

# Add vmls as alias to root's bashrc
echo 'alias vmls="hoster vm list"' >>/root/.bashrc

#_ GENERATE MINIMAL REQUIRED CONFIG FILES _#
mkdir -p ${HOSTER_WD}config_files/

### NETWORK CONFIG ###
cat <<EOF | cat >${HOSTER_WD}config_files/network_config.json
[
    {
        "network_name": "${NETWORK_NAME}",
        "network_gateway": "${NETWORK_BR_ADDR}",
        "network_subnet": "${NETWORK_SUBNET}",
        "network_range_start": "${NETWORK_RANGE_START}",
        "network_range_end": "${NETWORK_RANGE_END}",
        "bridge_interface": "None",
        "apply_bridge_address": true,
        "comment": "Internal Network"
    }
]
EOF

### HOST CONFIG ###
cat <<EOF | cat >${HOSTER_WD}config_files/host_config.json
{
    "public_vm_image_server": "https://images.yari.pw/",
    "active_datasets": [
        "zroot/vm-encrypted",
        "zroot/vm-unencrypted"
    ],
    "dns_servers": [
        "${UPSTREAM_DNS_SERVER}"
    ],
    "host_ssh_keys": [
        {
            "key_value": "${HOST_SSH_KEY}",
            "comment": "Host Key"
        }
    ]
}
EOF

#_ COPY OVER PF CONFIG _#
cat <<EOF | cat >/etc/pf.conf
table <private-ranges> { 10.0.0.0/8 100.64.0.0/10 127.0.0.0/8 169.254.0.0/16 172.16.0.0/12 192.0.0.0/24 192.0.0.0/29 192.0.2.0/24 192.88.99.0/24 192.168.0.0/16 198.18.0.0/15 198.51.100.0/24 203.0.113.0/24 240.0.0.0/4 255.255.255.255/32 }

set skip on lo0
scrub in all fragment reassemble max-mss 1440


### OUTBOUND NAT ###
nat on { ${PUBLIC_INTERFACE} } from { ${NETWORK_SUBNET} } to any -> { ${PUBLIC_INTERFACE} }


### INBOUND NAT EXAMPLES ###
# rdr pass on { ${PUBLIC_INTERFACE} } proto { tcp udp } from any to EXTERNAL_INTERFACE_IP_HERE port 80 -> 10.0.0.3 port 80                              # HTTP NAT Forwarding
# rdr pass on ${PUBLIC_INTERFACE} inet proto tcp from EXTERNAL_INTERFACE_IP_HERE to any port 80 -> EXTERNAL_INTERFACE_IP_HERE port 80                   # HTTP NAT Reflection

### ANTISPOOF RULE ###
antispoof quick for { ${PUBLIC_INTERFACE} } # DISABLE IF USING ANY ADDITIONAL ROUTERS IN THE VM, LIKE OPNSENSE


### FIREWALL RULES ###
# block in quick log on egress from <private-ranges>
# block return out quick on egress to <private-ranges>
block in all
pass out all keep state

# Allow internal NAT networks to go out + examples #
# pass in proto tcp to port 5900:5950 keep state
# pass in quick inet proto { tcp udp icmp } from { ${NETWORK_SUBNET} } to any                                                                           # Uncomment this rule to allow any traffic out
pass in quick inet proto { udp } from { ${NETWORK_SUBNET} } to { ${NETWORK_BR_ADDR} } port 53
block in quick inet from { ${NETWORK_SUBNET} } to <private-ranges>
pass in quick inet proto { tcp udp icmp } from { ${NETWORK_SUBNET} } to any


### INCOMING HOST RULES ###
pass in quick on { ${PUBLIC_INTERFACE} } inet proto icmp all # allow PING in
pass in quick on { ${PUBLIC_INTERFACE} } proto tcp to port 22 keep state #ALLOW_SSH_ACCESS_TO_HOST
# pass in proto tcp to port 80 keep state                                                                                                               # HTTP_NGINX_PROXY
# pass in proto tcp to port 443 keep state                                                                                                              # HTTPS_NGINX_PROXY
EOF

## SSH Banner
cat <<'EOF' | cat >/etc/motd.template
  ▄         ▄  ▄▄▄▄▄▄▄▄▄▄▄  ▄▄▄▄▄▄▄▄▄▄▄  ▄▄▄▄▄▄▄▄▄▄▄  ▄▄▄▄▄▄▄▄▄▄▄  ▄▄▄▄▄▄▄▄▄▄▄ 
 ▐░▌       ▐░▌▐░░░░░░░░░░░▌▐░░░░░░░░░░░▌▐░░░░░░░░░░░▌▐░░░░░░░░░░░▌▐░░░░░░░░░░░▌
 ▐░▌       ▐░▌▐░█▀▀▀▀▀▀▀█░▌▐░█▀▀▀▀▀▀▀▀▀  ▀▀▀▀█░█▀▀▀▀ ▐░█▀▀▀▀▀▀▀▀▀ ▐░█▀▀▀▀▀▀▀█░▌
 ▐░▌       ▐░▌▐░▌       ▐░▌▐░▌               ▐░▌     ▐░▌          ▐░▌       ▐░▌
 ▐░█▄▄▄▄▄▄▄█░▌▐░▌       ▐░▌▐░█▄▄▄▄▄▄▄▄▄      ▐░▌     ▐░█▄▄▄▄▄▄▄▄▄ ▐░█▄▄▄▄▄▄▄█░▌
 ▐░░░░░░░░░░░▌▐░▌       ▐░▌▐░░░░░░░░░░░▌     ▐░▌     ▐░░░░░░░░░░░▌▐░░░░░░░░░░░▌
 ▐░█▀▀▀▀▀▀▀█░▌▐░▌       ▐░▌ ▀▀▀▀▀▀▀▀▀█░▌     ▐░▌     ▐░█▀▀▀▀▀▀▀▀▀ ▐░█▀▀▀▀█░█▀▀ 
 ▐░▌       ▐░▌▐░▌       ▐░▌          ▐░▌     ▐░▌     ▐░▌          ▐░▌     ▐░▌  
 ▐░▌       ▐░▌▐░█▄▄▄▄▄▄▄█░▌ ▄▄▄▄▄▄▄▄▄█░▌     ▐░▌     ▐░█▄▄▄▄▄▄▄▄▄ ▐░▌      ▐░▌ 
 ▐░▌       ▐░▌▐░░░░░░░░░░░▌▐░░░░░░░░░░░▌     ▐░▌     ▐░░░░░░░░░░░▌▐░▌       ▐░▌
  ▀         ▀  ▀▀▀▀▀▀▀▀▀▀▀  ▀▀▀▀▀▀▀▀▀▀▀       ▀       ▀▀▀▀▀▀▀▀▀▀▀  ▀         ▀ 
      ┬  ┬┬┬─┐┌┬┐┬ ┬┌─┐┬  ┬┌─┐┌─┐┌┬┐┬┌─┐┌┐┌  ┌┬┐┌─┐┌┬┐┌─┐  ┌─┐┌─┐┌─┐┬ ┬  
      └┐┌┘│├┬┘ │ │ │├─┤│  │┌─┘├─┤ │ ││ ││││  │││├─┤ ││├┤   ├┤ ├─┤└─┐└┬┘  
       └┘ ┴┴└─ ┴ └─┘┴ ┴┴─┘┴└─┘┴ ┴ ┴ ┴└─┘┘└┘  ┴ ┴┴ ┴─┴┘└─┘  └─┘┴ ┴└─┘ ┴   


EOF
## EOF SSH Banner

wget https://github.com/yaroslav-gwit/HosterCore/releases/download/v0.3/hoster -O ${HOSTER_WD}hoster -q --show-progress
chmod +x ${HOSTER_WD}hoster

wget https://github.com/yaroslav-gwit/HosterCore/releases/download/v0.3/vm_supervisor_service -O ${HOSTER_WD}vm_supervisor_service -q --show-progress
chmod +x ${HOSTER_WD}vm_supervisor_service

wget https://github.com/yaroslav-gwit/HosterCore/releases/download/v0.3/self_update_service -O ${HOSTER_WD}self_update_service -q --show-progress
chmod +x ${HOSTER_WD}self_update_service

#_ LET USER KNOW THE STATE OF DEPLOYMENT _#
cat <<EOF | cat

╭────────────────────────────────────────────────────────────────────────────╮
│                                                                            │
│  The installation is now finished.                                         │
│  Your ZFS encryption password: it's right below this box                   │
│                                                                            │
│  Please save your password! If you lose it, your VMs on the encrypted      │
│  dataset will be lost!                                                     │
│                                                                            │
│  Reboot the system now to apply changes.                                   │
│                                                                            │
│  After the reboot mount the encrypted ZFS dataset and initialize Hoster    │
│  (these 2 steps are required after each reboot):                           │
│                                                                            │
│  zfs mount -a -l                                                           │
│  hoster init                                                               │
│                                                                            │
╰────────────────────────────────────────────────────────────────────────────╯
${ZFS_RANDOM_PASSWORD}

EOF
