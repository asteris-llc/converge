#!/usr/bin/env bash
set -eo pipefail

ROOT=$(pwd)

# we're going to spin up a bunch of servers. We need to make sure they're all
# torn down at the end.
PID=
finish() {
    kill -2 $PID
}
trap finish EXIT

# RPC Ping without SSL
"$ROOT"/converge server --no-token --log-level=ERROR &
PID=$!
sleep 0.5

if "$ROOT"/converge ping --log-level=ERROR > /dev/null; then
    echo "success: RPC ping without SSL"
else
    echo "FAIL: RPC ping without SSL"
    exit 1
fi

# HTTP Ping without SSL
if curl -fs http://localhost:4774/api/v1/ping > /dev/null; then
    echo "success: HTTP ping without SSL"
else
    echo "FAIL: HTTP ping without SSL"
    exit 1
fi

# restart with SSL
kill -2 $PID
sleep 0.5

"$ROOT"/converge server --log-level=ERROR \
                        --no-token \
                        --ca-file "$ROOT"/blackbox/ssl/Converge_Test.crt \
                        --cert-file "$ROOT"/blackbox/ssl/Converge_Test_Cert.crt \
                        --key-file "$ROOT"/blackbox/ssl/Converge_Test_Cert.key \
                        --use-ssl &
PID=$!
sleep 0.5

# RPC Ping with SSL
if "$ROOT"/converge ping --ca-file "$ROOT"/blackbox/ssl/Converge_Test.crt --use-ssl --log-level=ERROR > /dev/null; then
    echo "success: RPC ping with SSL"
else
    echo "FAIL: RPC ping with SSL"
    exit 1
fi

# HTTP Ping with SSL
if curl -fs --cacert "$ROOT"/blackbox/ssl/Converge_Test.crt https://localhost:4774/api/v1/ping > /dev/null; then
    echo "success: HTTP ping with SSL"
else
    echo "FAIL: HTTP ping with SSL"
    exit 1
fi
