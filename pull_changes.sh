#!/usr/bin/env sh
set -e

echo "=== Pulling changes from Git ==="
git stash
git pull
echo "=== Done pulling changes from Git ==="
