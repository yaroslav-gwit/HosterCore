#!/usr/bin/env sh
# set -e

bash pull_changes.sh
echo ""

bash build.sh
echo ""

echo "=== Starting the installation process ==="
cp -v hoster /opt/hoster-core/
cp -v vm_supervisor_service /opt/hoster-core/
cp -v self_update_service /opt/hoster-core/
cp -v dns_server /opt/hoster-core/
cp -v mbuffer/mbuffer /opt/hoster-core/
cp -v node_exporter/node_exporter /opt/hoster-core/node_exporter_custom
echo "Copied over executable files"

echo "=== Installation process done ==="
