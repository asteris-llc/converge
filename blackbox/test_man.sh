#!/usr/bin/env bash
set -eo pipefail

ROOT=$(pwd)

TMP=$(mktemp -d -t converge.man.XXXXXXXXXX)
function finish {
    rm -rf "$TMP"
}
trap finish EXIT

"$ROOT"/converge gen man --path="$TMP/man8"

for file in $TMP/man8/*; do
    page="$(basename "$file" | cut -d. -f1)"
    if man -M "$TMP" -w "$page" > /dev/null; then
        echo "found man page for $page"
    else
        echo "could not find man page for $page"
        exit 1
    fi
done
