#!/usr/bin/env bash
set -e
GO_BINARY=$(which go121)

# RED='\033[0;31m'
# LIGHT_RED='\033[1;31m'
GREEN='\033[0;32m'
LIGHT_GREEN='\033[1;32m'
NC='\033[0m'
# ERROR_TEXT="${LIGHT_RED}ERROR:${NC}"

echo -e "${LIGHT_GREEN}=== Starting the build process ===${NC}"
# Set the RELEASE=true, to build the release version
GIT_INFO=$(git describe --tags)
DATE_INFO=$(date '+%Y%m%d_%H%M%S')
DEV_VERSION=${GIT_INFO}_COMPTIME_${DATE_INFO}
RELEASE_VERSION=$(git describe --tags | sed 's/-[^-]*$//')

if test -z "${RELEASE}"; then
    echo -e "${GREEN}Building the DEV version of HosterCore${NC}"
    # echo ""
    # shellcheck disable=SC2059
    printf "Building the ${GREEN}hoster${NC} module ... "
    # go build -a -ldflags="-X HosterCore/cmd.HosterVersion=${VERSION}" -o hoster
    $GO_BINARY build -ldflags="-X HosterCore/cmd.HosterVersion=${DEV_VERSION}" -trimpath -o hoster
else
    echo -e "${GREEN}Building the RELEASE version of HosterCore${NC}"
    # echo ""
    # shellcheck disable=SC2059
    printf "Building the ${GREEN}hoster${NC} module ... "
    # $GO_BINARY build -trimpath -o hoster
    $GO_BINARY build -ldflags="-X HosterCore/cmd.HosterVersion=${RELEASE_VERSION}" -trimpath -o hoster
fi
# shellcheck disable=SC2059
printf "${GREEN}Done${NC}\n"

cd internal/app
# shellcheck disable=SC2059
printf "Building the ${GREEN}vm_supervisor_service${NC} module ... "
cd vm_supervisor/
# $GO_BINARY build -trimpath
if test -z "${RELEASE}"; then
    $GO_BINARY build -ldflags="-X main.version=${DEV_VERSION}" -trimpath
else
    $GO_BINARY build -ldflags="-X main.version=${RELEASE_VERSION}" -trimpath
fi
# shellcheck disable=SC2059
printf "${GREEN}Done${NC}\n"

# shellcheck disable=SC2059
printf "Building the ${GREEN}self_update${NC} module ... "
cd ..
cd self_update/
if test -z "${RELEASE}"; then
    $GO_BINARY build -ldflags="-X main.version=${DEV_VERSION}" -trimpath
else
    $GO_BINARY build -ldflags="-X main.version=${RELEASE_VERSION}" -trimpath
fi
# shellcheck disable=SC2059
printf "${GREEN}Done${NC}\n"

# shellcheck disable=SC2059
printf "Building the ${GREEN}dns_server${NC} module ... "
cd ..
cd dns_server/
# $GO_BINARY build -trimpath
if test -z "${RELEASE}"; then
    $GO_BINARY build -ldflags="-X main.version=${DEV_VERSION}" -trimpath
else
    $GO_BINARY build -ldflags="-X main.version=${RELEASE_VERSION}" -trimpath
fi
# shellcheck disable=SC2059
printf "${GREEN}Done${NC}\n"

# shellcheck disable=SC2059
printf "Building the ${GREEN}mbuffer${NC} limiter module ... "
cd ..
cd mbuffer/
# $GO_BINARY build -trimpath
if test -z "${RELEASE}"; then
    $GO_BINARY build -ldflags="-X main.version=${DEV_VERSION}" -trimpath
else
    $GO_BINARY build -ldflags="-X main.version=${RELEASE_VERSION}" -trimpath
fi
# shellcheck disable=SC2059
printf "${GREEN}Done${NC}\n"

# shellcheck disable=SC2059
printf "Building the ${GREEN}node_exporter_custom${NC} module ... "
cd ..
cd node_exporter/
# $GO_BINARY build -trimpath
if test -z "${RELEASE}"; then
    $GO_BINARY build -ldflags="-X main.version=${DEV_VERSION}" -trimpath
else
    $GO_BINARY build -ldflags="-X main.version=${RELEASE_VERSION}" -trimpath
fi
# shellcheck disable=SC2059
printf "${GREEN}Done${NC}\n"

# Legacy API v1
# printf "Building the ${GREEN}hoster_rest_api${NC} module ... "
# cd ..
# cd rest_api/
# go build -trimpath
# printf "${GREEN}Done${NC}\n"

# shellcheck disable=SC2059
printf "Building the ${GREEN}ha_watchdog${NC} module ... "
cd ..
cd ha_watchdog/
# $GO_BINARY build -trimpath
if test -z "${RELEASE}"; then
    $GO_BINARY build -ldflags="-X main.version=${DEV_VERSION}" -trimpath
else
    $GO_BINARY build -ldflags="-X main.version=${RELEASE_VERSION}" -trimpath
fi
# shellcheck disable=SC2059
printf "${GREEN}Done${NC}\n"

# shellcheck disable=SC2059
printf "Building the ${GREEN}scheduler${NC} module ... "
cd ..
cd scheduler/
# $GO_BINARY build -trimpath
if test -z "${RELEASE}"; then
    $GO_BINARY build -ldflags="-X main.version=${DEV_VERSION}" -trimpath
else
    $GO_BINARY build -ldflags="-X main.version=${RELEASE_VERSION}" -trimpath
fi
# shellcheck disable=SC2059
printf "${GREEN}Done${NC}\n"

# shellcheck disable=SC2059
printf "Building the ${GREEN}rest_api_v2${NC} module ... "
cd ..
cd rest_api_v2/
# $GO_BINARY build -trimpath
if test -z "${RELEASE}"; then
    $GO_BINARY build -ldflags="-X main.version=${DEV_VERSION}" -trimpath -o ./rest_api_v2
else
    $GO_BINARY build -ldflags="-X main.version=${RELEASE_VERSION}" -trimpath -o ./rest_api_v2
fi
# shellcheck disable=SC2059
printf "${GREEN}Done${NC}\n"

echo -e "${LIGHT_GREEN}=== Build process done ===${NC}"
