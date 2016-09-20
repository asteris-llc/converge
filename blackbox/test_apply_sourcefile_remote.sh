#!/usr/bin/env bash
set -eo pipefail

ROOT=$(pwd)
SOURCE=${1:-http://localhost:2694/api/v1/resources/modules/sourceFile.hcl}

${ROOT}/converge server --no-token --root ${ROOT}/samples &
PID=$!
function finish {
    kill -2 $PID
}
trap finish EXIT

sleep 0.5

${ROOT}/blackbox/test_apply.sh ${SOURCE}
