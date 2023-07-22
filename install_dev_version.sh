#!/usr/bin/env sh
# set -e

bash pull_changes.sh
echo ""

bash build.sh
echo ""

echo "=== Starting the installation process ==="
cp -v hoster /opt/hoster-core/
cp -v vm_supervisor/vm_supervisor /opt/hoster-core/vm_supervisor_service
cp -v self_update/self_update /opt/hoster-core/self_update_service
cp -v dns_server/dns_server /opt/hoster-core/
cp -v mbuffer/mbuffer /opt/hoster-core/
cp -v node_exporter/node_exporter /opt/hoster-core/node_exporter_custom
cp -v rest_api/rest_api /opt/hoster-core/hoster_rest_api
echo "Copied over executable files"

echo "=== Installation process done ==="
