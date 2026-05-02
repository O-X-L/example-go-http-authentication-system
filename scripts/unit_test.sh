#!/usr/bin/env bash

LIMIT="*"
if [[ -n "$1" ]]
then
  LIMIT="$1"
fi

set -euo pipefail

cd "$(dirname "$0")/.."

export APP_PEPPER=dsklkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkk
export APP_DEV=1

BASE_DIR="$(pwd)"

echo ''
echo '### API ###'
echo ''
cd "${BASE_DIR}/api"
go run gotest.tools/gotestsum@latest --format pkgname ./...

echo ''
echo '### FRONTEND SERVER ###'
echo ''
cd "${BASE_DIR}/frontend/server"
go run gotest.tools/gotestsum@latest --format pkgname ./...
