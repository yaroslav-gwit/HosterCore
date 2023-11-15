#!/usr/bin/env bash
# set -e

export BUILD_ALL=yes
if [[ -n $1 && $1 != all ]]; then BUILD_ALL=no; fi

# RED='\033[0;31m'
LIGHT_RED='\033[1;31m'
# GREEN='\033[0;32m'
LIGHT_GREEN='\033[1;32m'
NC='\033[0m'

ERROR_TEXT="${LIGHT_RED}ERROR:${NC}"

bash pull_changes.sh
echo ""

if ! bash build.sh; then echo -e "${ERROR_TEXT} Build process failed!" && exit 1; fi
echo ""

echo -e "${LIGHT_GREEN}=== Starting the installation process ===${NC}"

# 1
if [[ $1 == 1 || $1 == "hoster" || ${BUILD_ALL} == yes ]]; then
    cp -v hoster /opt/hoster-core/ 2>/dev/null || echo -e "${ERROR_TEXT} hoster binary is in use"
fi

# 2
if [[ $1 == 2 || $1 == "vm_supervisor" || ${BUILD_ALL} == yes ]]; then
    cp -v vm_supervisor/vm_supervisor /opt/hoster-core/vm_supervisor_service 2>/dev/null || echo -e "${ERROR_TEXT} vm_supervisor_service binary is in use"
fi

# 3
if [[ $1 == 3 || ${BUILD_ALL} == yes ]]; then
    cp -v self_update/self_update /opt/hoster-core/self_update_service 2>/dev/null || echo -e "${ERROR_TEXT} self_update_service binary is in use"
fi

# 4
if [[ $1 == 4 || ${BUILD_ALL} == yes ]]; then
    cp -v dns_server/dns_server /opt/hoster-core/ 2>/dev/null || echo -e "${ERROR_TEXT} dns_server binary is in use"
fi

# 5
if [[ $1 == 5 || ${BUILD_ALL} == yes ]]; then
    cp -v mbuffer/mbuffer /opt/hoster-core/ 2>/dev/null || echo -e "${ERROR_TEXT} mbuffer binary is in use"
fi

# 6
if [[ $1 == 6 || ${BUILD_ALL} == yes ]]; then
    cp -v node_exporter/node_exporter /opt/hoster-core/node_exporter_custom 2>/dev/null || echo -e "${ERROR_TEXT} node_exporter_custom binary is in use"
fi

# 7
if [[ $1 == 7 || ${BUILD_ALL} == yes ]]; then
    cp -v rest_api/rest_api /opt/hoster-core/hoster_rest_api 2>/dev/null || echo -e "${ERROR_TEXT} hoster_rest_api binary is in use"
fi

# 8
if [[ $1 == 8 || ${BUILD_ALL} == yes ]]; then
    cp -v ha_watchdog/ha_watchdog /opt/hoster-core/ 2>/dev/null || echo -e "${ERROR_TEXT} ha_watchdog binary is in use"
fi

echo -e "${LIGHT_GREEN}==== Copied over executable files ====${NC}"
echo -e "${LIGHT_GREEN}=== Installation process done ===${NC}"
