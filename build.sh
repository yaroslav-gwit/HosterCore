#!/usr/bin/env sh
echo "=== Starting the build process ==="
set -e

GIT_INFO=$(git describe --tags); DATE_INFO=$(date '+%Y-%m-%d_%H-%M-%S')
VERSION=${GIT_INFO}_COMPILED_ON_${DATE_INFO}  #; _=${VERSION}
# Set the RELEASE=true, to build the release version
if test -z "${RELEASE}"; then 
    echo "Building the DEV version of HosterCore..."
    echo ""

    echo "Building the hoster module..."
    # go build -a -ldflags="-X HosterCore/cmd.HosterVersion=${VERSION}" -o hoster
    go build -ldflags="-X HosterCore/cmd.HosterVersion=${VERSION}" -o hoster
else 
    echo "Building the RELEASE version of HosterCore..."
    echo ""

    echo "Building the hoster module..."
    go build -o hoster
fi
echo "Done"

echo "Building the vm_supervisor_service module..."
cd vm_supervisor/
go build
echo "Done"

echo "Building the self_update_service module..."
cd ..
cd self_update/
go build
echo "Done"

echo "Building the dns_server module..."
cd ..
cd dns_server/
go build 
echo "Done"

echo "Building the mbuffer limiter module..."
cd ..
cd mbuffer/
go build 
echo "Done"

echo "Building the node_exporter_custom module..."
cd ..
cd node_exporter/
go build 
echo "Done"

echo "Building the hoster_rest_api module..."
cd ..
cd rest_api/
go build 
echo "Done"

echo "Building the ha_watchdog module..."
cd ..
cd ha_watchdog/
go build 
echo "Done"

echo "=== Build process done ==="
