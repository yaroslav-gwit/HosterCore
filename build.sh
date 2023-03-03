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

echo "=== Build process done ==="
