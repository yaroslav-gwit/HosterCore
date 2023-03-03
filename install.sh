#!/usr/bin/env sh
set -e

bash pull_changes.sh
echo ""

bash build.sh
echo ""
echo "=== Starting the installation process ==="

mkdir -p /opt/hoster-core/
echo "Created /opt/hoster-core/ folder"

cp hoster /opt/hoster-core/
cp vm_supervisor_service /opt/hoster-core/
echo "Copied over executable files"

cp -r config_files /opt/hoster-core/
echo "Copied over config files"

rm -f /bin/hoster
ln -s /opt/hoster-core/hoster /bin/hoster
echo "Linked 'hoster' to '/bin/hoster' for system-wide use"

echo "=== Installation process done ==="
