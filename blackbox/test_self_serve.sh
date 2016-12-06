#!/usr/bin/env bash
set -eo pipefail

ROOT=$(pwd)

"$ROOT"/converge server --root "$ROOT"/samples --self-serve --no-token &
PID=$!
function finish {
    kill -2 $PID
}
trap finish EXIT

sleep 0.5

REMOTE_SUM=$(curl http://localhost:4774/api/v1/resources/binary -H "Accept: text/plain" | shasum | awk '{ print $1 }')
LOCAL_SUM=$(shasum "$ROOT"/converge | awk '{ print $1 }')

if [[ "$REMOTE_SUM" == "$LOCAL_SUM" ]]; then
    echo "Remote and local sums match"
else
    echo "Sums did not match! $REMOTE_SUM != $LOCAL_SUM"
    exit 1
fi
