#!/usr/bin/env sh
set -e

bash pull_changes.sh
echo ""

bash build.sh
echo ""
echo "=== Starting the installation process ==="

cp hoster /opt/hoster-core/
cp vm_supervisor_service /opt/hoster-core/
cp self_update_service /opt/hoster-core/
cp dns_server /opt/hoster-core/
echo "Copied over executable files"

echo "=== Installation process done ==="
