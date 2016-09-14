#!/usr/bin/env bash
set -eo pipefail

ROOT=$(pwd)
SOURCE=${1:-${ROOT}/samples/basic.hcl}
KEY=${1:-${ROOT}/samples/pubkeys.gpg}

TMP=$(mktemp -d -t converge.verify.XXXXXXXXXX)
function finish {
    rm -rf $TMP
}
trap finish EXIT

pushd $TMP

mkdir trustedkeys
cp $KEY trustedkeys/74fdf669f18d59f92b0aaccd720351ff475cc928

$ROOT/converge plan --local --verify-modules $SOURCE

echo "success: no errors during module verification"

popd
