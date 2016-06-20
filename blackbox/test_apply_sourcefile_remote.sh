#!/usr/bin/env bash
set -eo pipefail

ROOT=$(pwd)
SOURCE=${1:-http://localhost:8080/modules/sourceFile.hcl}

$ROOT/converge server --root $ROOT/samples &
PID=$!

$ROOT/blackbox/test_apply.sh $SOURCE

kill -2 $PID
