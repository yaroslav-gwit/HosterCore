#!/usr/bin/env sh
echo "=== Starting the build process ==="
set -e

echo "Building the hoster module..."
go build
echo "Done"

echo "Building the vm_supervisor module..."
cd vm_supervisor/
go build
echo "Done"

echo "Building the self_update module..."
cd ..
cd self_update/
go build
echo "Done"

echo "Building the dns_server module..."
cd ..
cd dns_server/
go build 
echo "Done"

echo "Building the mbuffer limiter..."
cd ..
cd mbuffer/
go build 
echo "Done"

echo "Building the node_exporter (hoster specific version)..."
cd ..
cd node_exporter/
go build 
echo "Done"

echo "Building the REST API module..."
cd ..
cd rest_api/
go build 
echo "Done"

echo "=== Build process done ==="
