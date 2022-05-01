#!/usr/bin/env bash

set -euo pipefail

SCRIPT_DIRECTORY="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
ROOT_DIRECTORY="$(dirname "$SCRIPT_DIRECTORY")"

pushd "$ROOT_DIRECTORY" > /dev/null 2>&1

./scripts/protobuf_build.sh

go build -o ./dist/client ./cmd/client.go
go build -o ./dist/server ./cmd/server.go

popd > /dev/null 2>&1