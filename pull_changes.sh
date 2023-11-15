#!/usr/bin/env bash
set -e

# RED='\033[0;31m'
LIGHT_RED='\033[1;31m'
# GREEN='\033[0;32m'
LIGHT_GREEN='\033[1;32m'
NC='\033[0m'

ERROR_TEXT="${LIGHT_RED}ERROR:${NC}"

echo "${LIGHT_GREEN}=== Pulling changes from Git ===${NC}"
git stash
git pull || echo -e "${ERROR_TEXT} could not pull from the Git repo"
echo "${LIGHT_GREEN}=== Done pulling changes from Git ===${NC}"
