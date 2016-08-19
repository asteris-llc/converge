#!/usr/bin/env bash
set -eo pipefail

ROOT=$(pwd)

TMP=$(mktemp -d -t converge.graph_dependencies.XXXXXXXXXX)
function finish {
    rm -rf $TMP
}
trap finish EXIT

pushd $TMP

for src in ${ROOT}/samples/dependencies/*.hcl; do
	b=$(basename $src)
	echo "starting repetative graph test for ${b}"
	for i in `seq 1 5`; do 
		${ROOT}/converge graph -l "WARN" ${src} >/dev/null
	done
	echo "success: no errors generating dependencies for ${b}"
done
echo "success: no errors for any .hcl files"

popd
