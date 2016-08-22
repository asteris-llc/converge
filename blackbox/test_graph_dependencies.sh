#!/usr/bin/env bash
set -eo pipefail

ROOT=$(pwd)

TMP=$(mktemp -d -t converge.graph_dependencies.XXXXXXXXXX)
function finish {
    rm -rf $TMP
}
trap finish EXIT

pushd $TMP

for src in ${ROOT}/samples/testdata/*.hcl; do
	b=$(basename $src)
	# since panic doesn't happen 100% of the time, run test a few times
	for i in `seq 1 5`; do 
		${ROOT}/converge graph -l "WARN" ${src} >/dev/null
	done
	echo "success: no errors generating dependencies for ${b}"
done
echo "success: no errors for any .hcl files"

popd
