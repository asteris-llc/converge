#!/usr/bin/env bash
set -eo pipefail

ROOT=$(pwd)
SOURCE=${1:-http://localhost:8080/modules/sourceFile.hcl}

$ROOT/converge server --root $ROOT/samples &
PID=$!
function finish {
    kill -2 $PID
}
trap finish EXIT

$ROOT/blackbox/test_apply.sh $SOURCE
