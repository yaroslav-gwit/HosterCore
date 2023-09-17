#!/usr/bin/env sh
# set -e

bash pull_changes.sh
echo ""

bash build.sh
echo ""

echo "=== Starting the installation process ==="
cp -v hoster /opt/hoster-core/ 2>/dev/null || echo "ERROR: hoster binary is in use"
cp -v vm_supervisor/vm_supervisor /opt/hoster-core/vm_supervisor_service 2>/dev/null || echo "ERROR: vm_supervisor_service binary is in use"
cp -v self_update/self_update /opt/hoster-core/self_update_service 2>/dev/null || echo "ERROR: self_update_service binary is in use"
cp -v dns_server/dns_server /opt/hoster-core/ 2>/dev/null || echo "ERROR: dns_server binary is in use"
cp -v mbuffer/mbuffer /opt/hoster-core/ 2>/dev/null || echo "ERROR: mbuffer binary is in use"
cp -v node_exporter/node_exporter /opt/hoster-core/node_exporter_custom 2>/dev/null || echo "ERROR: node_exporter_custom binary is in use"
cp -v rest_api/rest_api /opt/hoster-core/hoster_rest_api 2>/dev/null || echo "ERROR: hoster_rest_api binary is in use"
cp -v ha_watchdog/ha_watchdog /opt/hoster-core/ 2>/dev/null || echo "ERROR: ha_watchdog binary is in use"
echo "Copied over executable files"

echo "=== Installation process done ==="
