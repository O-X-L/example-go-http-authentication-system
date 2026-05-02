#!/usr/bin/env bash

set -euo pipefail

if ! which docker >/dev/null
then
  echo "ERROR: Docker required!"
  exit 1
fi

docker run -it --rm --name example-dev-db -e POSTGRES_USER=example -e POSTGRES_PASSWORD=dev-secret -e POSTGRES_DB=example -p 127.0.0.1:5432:5432 postgres
