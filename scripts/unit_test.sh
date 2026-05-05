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
export APP_SMTP_EMAIL=xxx@test.oxl.app
export APP_SMTP_SERVER=nosrv.test.oxl.app
export APP_SMTP_USER=test
export APP_SMTP_PASSWORD=xxx
export APP_GOOGLE_OAUTH_CLIENT_ID=xxx.apps.googleusercontent.com

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
