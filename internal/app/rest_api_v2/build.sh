#!/usr/bin/env bash
# This is a temporary build file for the separate APIv2 module.
# Don't forget to execute `swag init --pd` to generate the docs prior to the pull.

GO_BINARY=$(which go121)

git pull
$GO_BINARY build -trimpath -o ./rest_api_v2

./rest_api_v2
