#!/usr/bin/env bash

#_ CHECK IF USER IS ROOT _#
if [ "$EUID" -ne 0 ]; then echo " 🚦 ERROR: Please run this script as root user!" && exit 1; fi
#_ EOF CHECK IF USER IS ROOT _#

#_ SET DEFAULT VARS _#
NETWORK_NAME="${DEF_NETWORK_NAME:=internal}"
NETWORK_BR_ADDR="${DEF_NETWORK_BR_ADDR:=10.0.101.254}"
NETWORK_SUBNET="${DEF_NETWORK_SUBNET:=10.0.101.0/24}"
NETWORK_RANGE_START="${DEF_NETWORK_RANGE_START:=10.0.101.10}"
NETWORK_RANGE_END="${DEF_NETWORK_RANGE_END:=10.0.101.200}"
PUBLIC_INTERFACE="${DEF_PUBLIC_INTERFACE:=$(ifconfig | head -1 | awk '{ print $1 }' | sed s/://)}"
UPSTREAM_DNS_SERVER="${DEF_UPSTREAM_DNS_SERVER:=1.1.1.2}"
#_ EOF SET DEFAULT VARS _#

#_ CREATE AND SET A WORKING DIRECTORY _#
zfs create zroot/opt
zfs set mountpoint=/opt zroot/opt
zfs mount -a
mkdir /opt/hoster-core
HOSTER_WD="/opt/hoster-core/"
#_ EOF CREATE AND SET A WORKING DIRECTORY _#

#_ INSTALL THE REQUIRED PACKAGES _#
pkg update
pkg upgrade -y
pkg install -y vim bash bash-completion pftop tmux qemu-tools git curl
pkg install -y bhyve-firmware uefi-edk2-bhyve-csm edk2-bhyve openssl
pkg install -y htop wget gtar unzip cdrkit-genisoimage go121 beadm chrony
pkg install -y exa bat micro # modern alternatives to `cat` and `ls`
#_ EOF INSTALL THE REQUIRED PACKAGES _#

#_ OPTIONAL PACKAGES _#
# (install for easier debugging)
# pkg install -y nano micro bmon iftop mc fusefs-sshfs gnu-watch fping fish bhyve-rc grub2-bhyve

# Enable Chrony as a main source of time, and disable the old `ntpd` and `ntpdate`
service chronyd enable
(
    service ntpd stop
    service ntpdate stop
) || true
(
    service ntpd disable
    service ntpdate disable
) || true
service chronyd start
# EOF Enable Chrony as a main source of time, and disable the old `ntpd` and `ntpdate`

# Link bash into /bin/bash for better discover-ability
if [[ -f /bin/bash ]]; then rm /bin/bash; fi
ln -s "$(which bash)" /bin/bash
# EOF Link bash into /bin/bash for better discover-ability

#_ Set the ZFS encryption password _#
if [ -z "${DEF_ZFS_ENCRYPTION_PASSWORD}" ]; then
    ZFS_RANDOM_PASSWORD=$(openssl rand -base64 32 | tr -dc '[:alnum:]')
else
    ZFS_RANDOM_PASSWORD=${DEF_ZFS_ENCRYPTION_PASSWORD}
fi
#_ EOF Set the ZFS encryption password _#

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
#_ EOF GENERATE SSH KEYS _#

#_ REGISTER IF THE REQUIRED DATASETS EXIST _#
ENCRYPTED_DS=$(zfs list | grep -c "zroot/vm-encrypted")
UNENCRYPTED_DS=$(zfs list | grep -c "zroot/vm-unencrypted")
#_ EOF REGISTER IF THE REQUIRED DATASETS EXIST _#

#_ CREATE ZFS DATASETS IF THEY DON'T EXIST _#
if [[ ${ENCRYPTED_DS} -lt 1 ]]; then
    zpool set autoexpand=on zroot
    zpool set autoreplace=on zroot
    # zfs set primarycache=metadata zroot
    echo -e "${ZFS_RANDOM_PASSWORD}" | zfs create -o encryption=on -o keyformat=passphrase zroot/vm-encrypted
fi

if [[ ${UNENCRYPTED_DS} -lt 1 ]]; then
    zpool set autoexpand=on zroot
    zpool set autoreplace=on zroot
    # zfs set primarycache=metadata zroot
    zfs create zroot/vm-unencrypted
fi
#_ EOF CREATE ZFS DATASETS IF THEY DON'T EXIST _#

#_ BOOTLOADER OPTIMIZATIONS _#
BOOTLOADER_FILE="/boot/loader.conf"
## Deprecated values, will be removed in the next release
# CMD_LINE='fusefs_load="YES"' && if [[ $(grep -c "${CMD_LINE}" ${BOOTLOADER_FILE}) -lt 1 ]]; then echo "${CMD_LINE}" >> ${BOOTLOADER_FILE}; fi
# CMD_LINE='vm.kmem_size="400M"' && if [[ $(grep -c "${CMD_LINE}" ${BOOTLOADER_FILE}) -lt 1 ]]; then echo "${CMD_LINE}" >>${BOOTLOADER_FILE}; fi
# CMD_LINE='vm.kmem_size_max="400M"' && if [[ $(grep -c "${CMD_LINE}" ${BOOTLOADER_FILE}) -lt 1 ]]; then echo "${CMD_LINE}" >>${BOOTLOADER_FILE}; fi
# CMD_LINE='vfs.zfs.arc_max="40M"' && if [[ $(grep -c "${CMD_LINE}" ${BOOTLOADER_FILE}) -lt 1 ]]; then echo "${CMD_LINE}" >>${BOOTLOADER_FILE}; fi
# CMD_LINE='vfs.zfs.vdev.cache.size="5M"' && if [[ $(grep -c "${CMD_LINE}" ${BOOTLOADER_FILE}) -lt 1 ]]; then echo "${CMD_LINE}" >>${BOOTLOADER_FILE}; fi
# CMD_LINE='virtio_blk_load="YES"' && if [[ $(grep -c "${CMD_LINE}" ${BOOTLOADER_FILE}) -lt 1 ]]; then echo "${CMD_LINE}" >> ${BOOTLOADER_FILE}; fi
## Up-to-date values
CMD_LINE='# vfs.zfs.arc.max=367001600  # 350MB -> Min possible ZFS ARC Limit on FreeBSD' && if [[ $(grep -c "${CMD_LINE}" ${BOOTLOADER_FILE}) -lt 1 ]]; then echo "${CMD_LINE}" >>${BOOTLOADER_FILE}; fi
CMD_LINE='vfs.zfs.arc.max=1073741824  # 1G ZFS ARC Limit' && if [[ $(grep -c "${CMD_LINE}" ${BOOTLOADER_FILE}) -lt 1 ]]; then echo "${CMD_LINE}" >>${BOOTLOADER_FILE}; fi
CMD_LINE='pf_load="YES"' && if [[ $(grep -c "${CMD_LINE}" ${BOOTLOADER_FILE}) -lt 1 ]]; then echo "${CMD_LINE}" >>${BOOTLOADER_FILE}; fi
CMD_LINE='kern.racct.enable=1' && if [[ $(grep -c "${CMD_LINE}" ${BOOTLOADER_FILE}) -lt 1 ]]; then echo "${CMD_LINE}" >>${BOOTLOADER_FILE}; fi
# Install a better (official) Realtek driver to improve the stability and performance
ifconfig re0 &>/dev/null && echo " 🔷 DEBUG: Realtek interface detected, installing realtek-re-kmod driver and enabling boot time optimizations for it"
ifconfig re0 &>/dev/null && pkg install -y realtek-re-kmod
ifconfig re0 &>/dev/null && CMD_LINE='if_re_load="YES"' && if [[ $(grep -c "${CMD_LINE}" ${BOOTLOADER_FILE}) -lt 1 ]]; then echo "${CMD_LINE}" >>${BOOTLOADER_FILE}; fi
ifconfig re0 &>/dev/null && CMD_LINE='if_re_name="/boot/modules/if_re.ko"' && if [[ $(grep -c "${CMD_LINE}" ${BOOTLOADER_FILE}) -lt 1 ]]; then echo "${CMD_LINE}" >>${BOOTLOADER_FILE}; fi
ifconfig re0 &>/dev/null && CMD_LINE='# Disable the below if you are using Jumbo frames' && if [[ $(grep -c "${CMD_LINE}" ${BOOTLOADER_FILE}) -lt 1 ]]; then echo "${CMD_LINE}" >>${BOOTLOADER_FILE}; fi
ifconfig re0 &>/dev/null && CMD_LINE='hw.re.max_rx_mbuf_sz="2048"' && if [[ $(grep -c "${CMD_LINE}" ${BOOTLOADER_FILE}) -lt 1 ]]; then echo "${CMD_LINE}" >>${BOOTLOADER_FILE}; fi
#_ EOF BOOTLOADER OPTIMIZATIONS _#

#_ PF CONFIG BLOCK IN rc.conf _#
RC_CONF_FILE="/etc/rc.conf"
## Up-to-date values
CMD_LINE='pf_enable="yes"' && if [[ $(grep -c "${CMD_LINE}" ${RC_CONF_FILE}) -lt 1 ]]; then echo "${CMD_LINE}" >>${RC_CONF_FILE}; fi
CMD_LINE='pf_rules="/etc/pf.conf"' && if [[ $(grep -c "${CMD_LINE}" ${RC_CONF_FILE}) -lt 1 ]]; then echo "${CMD_LINE}" >>${RC_CONF_FILE}; fi
CMD_LINE='pflog_enable="yes"' && if [[ $(grep -c "${CMD_LINE}" ${RC_CONF_FILE}) -lt 1 ]]; then echo "${CMD_LINE}" >>${RC_CONF_FILE}; fi
CMD_LINE='pflog_logfile="/var/log/pflog"' && if [[ $(grep -c "${CMD_LINE}" ${RC_CONF_FILE}) -lt 1 ]]; then echo "${CMD_LINE}" >>${RC_CONF_FILE}; fi
CMD_LINE='pflog_flags=""' && if [[ $(grep -c "${CMD_LINE}" ${RC_CONF_FILE}) -lt 1 ]]; then echo "${CMD_LINE}" >>${RC_CONF_FILE}; fi
CMD_LINE='gateway_enable="yes"' && if [[ $(grep -c "${CMD_LINE}" ${RC_CONF_FILE}) -lt 1 ]]; then echo "${CMD_LINE}" >>${RC_CONF_FILE}; fi
#_ EOF PF CONFIG BLOCK IN rc.conf _#

# Set .profile for the `root` user
cat <<'EOF' | cat >/root/.profile
PATH=/sbin:/bin:/usr/sbin:/usr/bin:/usr/local/sbin:/usr/local/bin:~/bin:/opt/hoster-core; export PATH
HOME=/root; export HOME
TERM=${TERM:-xterm}; export TERM
PAGER=less; export PAGER
EDITOR=vim; export EDITOR

# set ENV to a file invoked each time sh is started for interactive use.
ENV=$HOME/.shrc; export ENV

# Query terminal size; useful for serial lines.
if [ -x /usr/bin/resizewin ] ; then /usr/bin/resizewin -z ; fi

# Display Hoster version on login
[ -z "$PS1" ] && true || echo "Hoster version: $(/opt/hoster-core/hoster version)"

# Add some common Hoster commands as aliases to type less
alias vms="hoster vm list"
alias vmsu="hoster vm list -u"
alias jails="hoster jail list"
alias jailsu="hoster jail list -u"

# Enable bash completion
[[ $PS1 && -f /usr/local/share/bash-completion/bash_completion.sh ]] && source /usr/local/share/bash-completion/bash_completion.sh
EOF
# EOF Set .profile for the `root` user

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

### REST API CONFIG ###
API_RANDOM_PASSWORD=$(openssl rand -base64 32 | tr -dc '[:alnum:]')
HA_RANDOM_PASSWORD=$(openssl rand -base64 32 | tr -dc '[:alnum:]')
cat <<EOF | cat >${HOSTER_WD}config_files/restapi_config.json
{
    "bind": "0.0.0.0",
    "port": 3000,
    "protocol": "http",
    "ha_mode": false,
    "ha_debug": true,
    "http_auth": [
        {
            "user": "admin",
            "password": "${API_RANDOM_PASSWORD}",
            "ha_user": false
        },
        {
            "user": "ha_user",
            "password": "${HA_RANDOM_PASSWORD}",
            "ha_user": true
        }
     ]
}
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
#_ EOF GENERATE MINIMAL REQUIRED CONFIG FILES _#

#_ COPY OVER PF CONFIG _#
cat <<EOF | cat >/etc/pf.conf
table <private-ranges> { 10.0.0.0/8 100.64.0.0/10 127.0.0.0/8 169.254.0.0/16 172.16.0.0/12 192.0.0.0/24 192.0.0.0/29 192.0.2.0/24 192.88.99.0/24 192.168.0.0/16 198.18.0.0/15 198.51.100.0/24 203.0.113.0/24 240.0.0.0/4 255.255.255.255/32 }

set skip on lo0
scrub in all fragment reassemble max-mss 1440

### OUTBOUND NAT ###
nat on { ${PUBLIC_INTERFACE} } from { ${NETWORK_SUBNET} } to any -> { ${PUBLIC_INTERFACE} }

### INBOUND NAT EXAMPLES ###
# rdr pass on { ${PUBLIC_INTERFACE} } proto { tcp } from any to EXTERNAL_INTERFACE_IP port 80 -> { VM or Jail name, or hardcoded IP address } port 80  # HTTP NAT Forwarding
# rdr pass on { vm-${NETWORK_NAME} } proto { tcp } from any to EXTERNAL_INTERFACE_IP port 80 -> { VM or Jail name, or hardcoded IP address } port 80  # HTTP RDR Reflection 

### ANTISPOOF RULE ###
antispoof quick for { ${PUBLIC_INTERFACE} }  # COMMENT OUT IF YOU USE ANY VM-based ROUTERS, like OPNSense, pfSense, etc.

### FIREWALL RULES ###
# block in quick log on egress from <private-ranges>
# block return out quick on egress to <private-ranges>
block in all
pass out all keep state

# Allow internal NAT networks to go out + examples #
# pass in proto tcp to port 5900:5950 keep state  # Allow access to VNC ports from any IP
# pass in quick inet proto { tcp udp icmp } from { ${NETWORK_SUBNET} } to any  # Uncomment this rule to allow any traffic out
pass in quick inet proto { udp } from { ${NETWORK_SUBNET} } to { ${NETWORK_BR_ADDR} } port 53  # Allow access to the internal DNS server
block in quick inet from { ${NETWORK_SUBNET} } to <private-ranges>  # Block access from the internal network
pass in quick inet proto { tcp udp icmp } from { ${NETWORK_SUBNET} } to any  # Together with the above rule allows access to only external resources

### INCOMING HOST RULES ###
pass in quick on { ${PUBLIC_INTERFACE} } inet proto icmp all  # Allow PING from any IP to this host
pass in quick on { ${PUBLIC_INTERFACE} } proto tcp to port 22 keep state  # Allow SSH from any IP to this host
# pass in proto tcp to port 80 keep state  # Allow access to internal Traefik service
# pass in proto tcp to port 443 keep state  # Allow access to internal Traefik service
EOF
#_ EOF COPY OVER PF CONFIG _#

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

# Download all Hoster-related binaries
wget https://github.com/yaroslav-gwit/HosterCore/releases/download/v0.3/hoster -O ${HOSTER_WD}hoster -q --show-progress
chmod 0755 ${HOSTER_WD}hoster
# TBD in the new release rename to vm_supervisor instead
wget https://github.com/yaroslav-gwit/HosterCore/releases/download/v0.3/vm_supervisor_service -O ${HOSTER_WD}vm_supervisor_service -q --show-progress
chmod 0755 ${HOSTER_WD}vm_supervisor_service
# EOF TBD in the new release rename to vm_supervisor instead
# TBD in the new release rename to self_update instead
wget https://github.com/yaroslav-gwit/HosterCore/releases/download/v0.3/self_update_service -O ${HOSTER_WD}self_update_service -q --show-progress
chmod 0755 ${HOSTER_WD}self_update_service
# EOF TBD in the new release rename to self_update instead
wget https://github.com/yaroslav-gwit/HosterCore/releases/download/v0.3/node_exporter_custom -O ${HOSTER_WD}node_exporter_custom -q --show-progress
chmod 0755 ${HOSTER_WD}node_exporter_custom
wget https://github.com/yaroslav-gwit/HosterCore/releases/download/v0.3/mbuffer -O ${HOSTER_WD}mbuffer -q --show-progress
chmod 0755 ${HOSTER_WD}mbuffer
wget https://github.com/yaroslav-gwit/HosterCore/releases/download/v0.3/hoster_rest_api -O ${HOSTER_WD}hoster_rest_api -q --show-progress
chmod 0755 ${HOSTER_WD}hoster_rest_api
wget https://github.com/yaroslav-gwit/HosterCore/releases/download/v0.3/ha_watchdog -O ${HOSTER_WD}ha_watchdog -q --show-progress
chmod 0755 ${HOSTER_WD}ha_watchdog
wget https://github.com/yaroslav-gwit/HosterCore/releases/download/v0.3/dns_server -O ${HOSTER_WD}dns_server -q --show-progress
chmod 0755 ${HOSTER_WD}dns_server
# EOF Download all Hoster-related binaries

# Enable basic bash completion
${HOSTER_WD}hoster completion bash >/usr/local/etc/bash_completion.d/hoster-completion.bash && echo " 🔷 DEBUG: Bash completion for Hoster has been enabled"
chmod 0755 /usr/local/etc/bash_completion.d/hoster-completion.bash
# EOF Enable basic bash completion

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
 !!! IMPORTANT !!! ZFS Encryption Password: ${ZFS_RANDOM_PASSWORD}

EOF
#_ EOF LET USER KNOW THE STATE OF DEPLOYMENT _#
