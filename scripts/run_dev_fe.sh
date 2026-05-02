#!/usr/bin/env bash

set -euo pipefail

cd "$(dirname "$0")/.."
BASE_DIR="$(pwd)"

echo '### BUILDING... ###'

if which npm >/dev/null
then
  cd "${BASE_DIR}/frontend/client"
  if ! [ -d "${BASE_DIR}/frontend/client/node_modules" ]
  then
    npm install
  fi
  npm run build
fi

bash "${BASE_DIR}/scripts/build.sh"

echo '### RUNNING ###'
export APP_DEV=1

${BASE_DIR}/build/fe-server
