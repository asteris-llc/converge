#!/usr/bin/env bash
set -eo pipefail

ROOT=$(pwd)
TMP=$(mktemp -d -t converge.graphviz.XXXXXXXXXX)

"$ROOT"/converge server --no-token &
PID=$!

function finish {
    rm -fr "$TMP"
    kill -2 "$PID"
}
trap finish EXIT

pushd "$TMP"

for i in "$ROOT"/samples/*.hcl; do
    b=$(basename "$i")
    dotSource="${b}.dot"
    pngOutput="${dotSource}.png"
    "$ROOT"/converge graph "$i" > "$dotSource"
    if [[ ! $? ]]; then
        echo "failed to generate graph for ${b}"
        exit 1
    fi
    dot -Tpng "$dotSource" -o "$pngOutput"
    if [[ ! $? ]]; then
        echo "dot failed on output from ${b}"
        exit 1
    fi
    echo "success: generated graph for ${b}"
done

echo "success: generated all graphs successfully"

popd
