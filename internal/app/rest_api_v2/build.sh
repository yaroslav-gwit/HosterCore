#!/usr/bin/env bash
# This is a temporary build file for the separate APIv2 module.
# Don't forget to execute `swag init --pd` to generate the docs prior to the pull.
git pull
go121 build -trimpath -o ./rest_api_v2
./rest_api_v2
