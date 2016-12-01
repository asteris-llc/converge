#!/usr/bin/env bash
set -eo pipefail

ROOT=$(pwd)
SOURCE=${1:-http://localhost:4774/api/v1/resources/modules/basic.hcl}

"$ROOT"/converge server --no-token --root "$ROOT"/samples &

PID=$!
function finish {
    kill -2 "$PID"
    rm -rf "$TMP"
}
trap finish EXIT

sleep 0.5

"$ROOT"/blackbox/test_apply.sh "$SOURCE" "$TMP"
