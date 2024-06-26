#!/usr/bin/env bash
set -e

WORKDIR=$(pwd)

# shellcheck disable=SC2164
cd ./internal/app/rest_api_v2
GOOS=freebsd swag init --pd

cd "${WORKDIR}"
