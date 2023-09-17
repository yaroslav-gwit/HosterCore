#!/usr/bin/env sh
# set -e

bash pull_changes.sh
echo ""

bash build.sh
echo ""

echo "=== Starting the installation process ==="
cp hoster /opt/hoster-core/ || echo "ERROR: hoster binary is in use"
cp vm_supervisor/vm_supervisor /opt/hoster-core/vm_supervisor_service || echo "ERROR: vm_supervisor_service binary is in use"
cp self_update/self_update /opt/hoster-core/self_update_service || echo "ERROR: self_update_service binary is in use"
cp dns_server/dns_server /opt/hoster-core/ || echo "ERROR: dns_server binary is in use"
cp mbuffer/mbuffer /opt/hoster-core/ || echo "ERROR: mbuffer binary is in use"
cp node_exporter/node_exporter /opt/hoster-core/node_exporter_custom || echo "ERROR: node_exporter_custom binary is in use"
cp rest_api/rest_api /opt/hoster-core/hoster_rest_api || echo "ERROR: hoster_rest_api binary is in use"
cp ha_watchdog/ha_watchdog /opt/hoster-core/ || echo "ERROR: ha_watchdog binary is in use"
echo "Copied over executable files"

echo "=== Installation process done ==="
