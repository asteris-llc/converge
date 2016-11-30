#!/usr/bin/env bash
set -e

ROOT=$(pwd)
TOKEN=$(uuid)

"$ROOT"/converge server --root "$ROOT"/samples --self-serve --rpc-token="$TOKEN" &
PID=$!
function finish {
    kill -2 $PID
}
trap finish EXIT

sleep 0.5

if "$ROOT"/converge plan samples/basic.hcl 2>&1 | grep -q "authorization not provided"; then
    echo "SUCCESS: auth failed when token not provided"
else
    echo "FAIL: auth seems to have succeeded. This is an error, since no token was provided."
    exit 1
fi

if "$ROOT"/converge plan --rpc-token "wrong" samples/basic.hcl 2>&1 | grep -q "signature is invalid"; then
    echo "SUCCESS: auth failed when wrong token provided"
else
    echo "FAIL: auth seems to have succeeded. This is an error, since the wrong token was provided."
    exit 1
fi

if "$ROOT"/converge plan --rpc-token "$TOKEN" samples/basic.hcl >/dev/null 2>&1; then
    echo "SUCCESS: auth succeeded when correct token provided"
else
    echo "FAIL: auth seems to have failed. This is an error, since the correct token was provided."
    exit 1
fi
