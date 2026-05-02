#!/usr/bin/env bash

set -euo pipefail

cd "$(dirname "$0")/.."
BASE_DIR="$(pwd)"

PATH_OUT="$(pwd)/build"
mkdir -p "$PATH_OUT"

echo ''
echo ' > API'
echo ''
cd "${BASE_DIR}/api"
go build -o "${PATH_OUT}/api" ./cmd/main.go

echo ''
echo ' > FRONTEND SERVER'
echo ''
cd "${BASE_DIR}/frontend/server"
go build -o "${PATH_OUT}/fe-server" ./cmd/main.go

echo ''
echo '### DONE ###'
echo ''

ls -l "$PATH_OUT"

echo ''
echo ''
