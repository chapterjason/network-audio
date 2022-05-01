#!/usr/bin/env bash

SCRIPT_DIRECTORY="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
ROOT_DIRECTORY="$(dirname "$SCRIPT_DIRECTORY")"
SRC_DIR="$ROOT_DIRECTORY/pkg/messages"
DIST_DIR="$ROOT_DIRECTORY/pkg"

protoc -I="$SRC_DIR" --go_out="$DIST_DIR" "$SRC_DIR/audio.proto"
protoc -I="$SRC_DIR" --go_out="$DIST_DIR" "$SRC_DIR/time.proto"
protoc -I="$SRC_DIR" --go_out="$DIST_DIR" "$SRC_DIR/latency.proto"
protoc -I="$SRC_DIR" --go_out="$DIST_DIR" "$SRC_DIR/command.proto"