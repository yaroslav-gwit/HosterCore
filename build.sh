#!/usr/bin/env sh
echo "=== Starting the build process ==="
set -e

echo "Building the hoster module..."
go build
echo "Done"

echo "Building the vm_supervisor module..."
cd vm_supervisor/
go build
mv vm_supervisor ../vm_supervisor_service
echo "Done"

echo "Building the self_update module..."
cd ..
cd self_update/
go build
mv self_update ../self_update_service
echo "Done"

echo "Building the dns_server module..."
cd ..
cd dns_server/
go build
mv dns_server ../dns_server
echo "Done"

echo "=== Build process done ==="
