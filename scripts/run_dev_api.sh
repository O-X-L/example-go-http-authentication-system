#!/usr/bin/env bash

if [ -z "$APP_SMTP_EMAIL" ]
then
  echo "ERROR: APP_SMTP_EMAIL ENV-Var missing!"
  exit 1
fi
if [ -z "$APP_SMTP_SERVER" ]
then
  echo "ERROR: APP_SMTP_SERVER ENV-Var missing!"
  exit 1
fi
if [ -z "$APP_SMTP_USER" ]
then
  echo "ERROR: APP_SMTP_USER ENV-Var missing!"
  exit 1
fi
if [ -z "$APP_SMTP_PASSWORD" ]
then
  echo "ERROR: APP_SMTP_PASSWORD ENV-Var missing!"
  exit 1
fi
if [ -z "$APP_GOOGLE_OAUTH_CLIENT_ID" ]
then
  echo "ERROR: APP_GOOGLE_OAUTH_CLIENT_ID ENV-Var missing!"
  exit 1
fi

set -euo pipefail

cd "$(dirname "$0")/.."
BASE_DIR="$(pwd)"

echo '### BUILDING... ###'
bash "${BASE_DIR}/scripts/build.sh"

echo '### RUNNING ###'
export APP_PEPPER=dsklkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkk
export APP_DEV=1

${BASE_DIR}/build/api
