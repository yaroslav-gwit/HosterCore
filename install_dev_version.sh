#!/usr/bin/env bash

# set -e
# go build -ldflags="-s -w" hello.go  ## hint: this will allow you to build the binary without the debug information (stripped binary)
# export BUILD_ALL=yes
# if [[ -n $1 && $1 != all ]]; then BUILD_ALL=no; fi

# RED='\033[0;31m'
# GREEN='\033[0;32m'
LIGHT_RED='\033[1;31m'
LIGHT_GREEN='\033[1;32m'
NC='\033[0m'

ERROR_TEXT="${LIGHT_RED}ERROR:${NC}"

bash pull_changes.sh
echo ""
if ! bash build.sh; then echo -e "${ERROR_TEXT} Build process failed!" && exit 1; fi
echo ""

echo -e "${LIGHT_GREEN}=== Stopping the additional services ===${NC}"
hoster node_exporter stop
hoster scheduler stop
hoster dns stop
hoster api stop
echo -e "${LIGHT_GREEN}=== Stopping the additional services: DONE ===${NC}"

echo -e "${LIGHT_GREEN}=== Starting the installation process ===${NC}"

cp -v hoster /opt/hoster-core/ 2>/dev/null || echo -e "${ERROR_TEXT} hoster binary is in use"
cp -v internal/app/vm_supervisor/vm_supervisor /opt/hoster-core/vm_supervisor_service 2>/dev/null || echo -e "${ERROR_TEXT} vm_supervisor_service binary is in use"
# cp -v internal/app/self_update/self_update /opt/hoster-core/self_update_service 2>/dev/null || echo -e "${ERROR_TEXT} self_update_service binary is in use"
cp -v internal/app/self_update/self_update /opt/hoster-core/ 2>/dev/null || echo -e "${ERROR_TEXT} self_update binary is in use"
cp -v internal/app/dns_server/dns_server /opt/hoster-core/ 2>/dev/null || echo -e "${ERROR_TEXT} dns_server binary is in use"
cp -v internal/app/mbuffer/mbuffer /opt/hoster-core/ 2>/dev/null || echo -e "${ERROR_TEXT} mbuffer binary is in use"
cp -v internal/app/node_exporter/node_exporter /opt/hoster-core/node_exporter_custom 2>/dev/null || echo -e "${ERROR_TEXT} node_exporter_custom binary is in use"
cp -v internal/app/scheduler/scheduler /opt/hoster-core/ 2>/dev/null || echo -e "${ERROR_TEXT} scheduler binary is in use"
cp -v internal/app/ha_watchdog/ha_watchdog /opt/hoster-core/ 2>/dev/null || echo -e "${ERROR_TEXT} ha_watchdog binary is in use"

# TEMPORARILY DISABLED DUE TO v2 MIGRATION, WILL BE REMOVED LATER
# cp -v internal/app/rest_api/rest_api /opt/hoster-core/hoster_rest_api 2>/dev/null || echo -e "${ERROR_TEXT} hoster_rest_api binary is in use"

# API v2 Service
mkdir -p /opt/hoster-api/
cp -rv internal/app/rest_api_v2/docs /opt/hoster-api/ 2>/dev/null || echo -e "${ERROR_TEXT} could not copy rest_api_v2 docs"
cp -v internal/app/rest_api_v2/rest_api_v2 /opt/hoster-api/ 2>/dev/null || echo -e "${ERROR_TEXT} rest_api_v2 binary is in use"
chmod 0750 /opt/hoster-api/
chmod 0750 /opt/hoster-api/docs
chmod 0750 /opt/hoster-api/rest_api_v2
chmod 0640 /opt/hoster-api/docs/*
#
# echo -e "${LIGHT_GREEN}=== Installation process: DONE ===${NC}"
echo -e "${LIGHT_GREEN}=== Starting the installation process: DONE ===${NC}"

echo ""
echo -e "${LIGHT_GREEN}=== Starting the additional services back up ===${NC}"
hoster node_exporter start
hoster scheduler start
hoster dns start
hoster api start
echo -e "${LIGHT_GREEN}=== Starting the additional services back up: DONE ===${NC}"
