#!/usr/bin/env bash
set -eo pipefail

ROOT=$(pwd)

$ROOT/converge server --root $ROOT/samples --self-serve &
PID=$!

sleep 0.5

REMOTE_SUM=$(curl http://localhost:8080/bootstrap/binary | shasum | awk '{ print $1 }')
LOCAL_SUM=$(shasum $ROOT/converge | awk '{ print $1 }')

if [[ "$REMOTE_SUM" == "$LOCAL_SUM" ]]; then
    echo "Remote and local sums match"
else
    echo "Sums did not match! $REMOTE_SUM != $LOCAL_SUM"
    exit 1
fi

kill -2 $PID
