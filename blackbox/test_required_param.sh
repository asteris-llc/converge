#!/usr/bin/env bash
set -eo pipefail

ROOT=$(pwd)

TMP=$(mktemp -d -t converge.apply.XXXXXXXXXX)
function finish {
    rm -rf ${TMP}
}
trap finish EXIT

pushd ${TMP}

echo 'param "test" {}' > required_param.hcl

if ${ROOT}/converge apply --local required_param.hcl; then
    echo "failed: apply without required param succeeded"
    exit 1
else
    echo "success: apply without required param failed"
fi

popd
