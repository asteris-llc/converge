#!/usr/bin/env bash
set -eo pipefail

ROOT=$(pwd)
SOURCE=${1:-"$ROOT"/samples/conditional.hcl}

TMP=$(mktemp -d -t converge.apply.XXXXXXXXXX)
function finish {
    rm -rf "$TMP"
}
trap finish EXIT

pushd "$TMP"

"$ROOT"/converge apply --local "$SOURCE"

if [ -f foo-file.txt ]; then
    echo "foo-file.txt shouldn't exist!"
    exit 1
fi

if [ -f foo-file-2.txt ]; then
    echo "foo-file-2.txt shouldn't exist!"
    exit 1
fi

"$ROOT"/converge apply --local -p "val=1" "$SOURCE"

if [ ! -f foo-file.txt ]; then
    echo "foo-file.txt doesn't exist!"
    exit 1
fi

if [ ! -f foo-file-2.txt ]; then
    echo "foo-file-2.txt doesn't exist!"
    exit 1
fi

if [[ "$(cat foo-file.txt)" != "foo1" ]]; then
    echo "foo-file.txt doesn't have the right content"
    echo "has '$(at foo-file.txt)', want 'foo1'"
    exit 1
fi

if [[ "$(cat foo-file-2.txt)" != "foo-file.txt" ]]; then
    echo "foo-file-2.txt doesn't have the right content"
    echo "has '$(at foo-file-2.txt)', want 'foo-file.txt'"
    exit 1
fi

SOURCE=${1:-"$ROOT"/samples/conditionalLanguages.hcl}

"$ROOT"/converge apply --local "$SOURCE"

if [ ! -f greeting.txt ]; then
    echo "greeting.txt doesn't exist!"
    exit 1
fi

if [[ "$(cat greeting.txt)" != "hello" ]]; then
    echo "greeting.txt doesn't have the right content"
    echo "has '$(at greeting.txt)', want 'hello'"
    exit 1
fi

"$ROOT"/converge apply --local -p "lang=spanish" "$SOURCE"

if [ ! -f greeting.txt ]; then
    echo "greeting.txt doesn't exist!"
    exit 1
fi

if [[ "$(cat greeting.txt)" != "hola" ]]; then
    echo "greeting.txt doesn't have the right content"
    echo "has '$(at greeting.txt)', want 'hola'"
    exit 1
fi

"$ROOT"/converge apply --local -p "lang=french" "$SOURCE"

if [ ! -f greeting.txt ]; then
    echo "greeting.txt doesn't exist!"
    exit 1
fi

if [[ "$(cat greeting.txt)" != "salut" ]]; then
    echo "greeting.txt doesn't have the right content"
    echo "has '$(at greeting.txt)', want 'salut'"
    exit 1
fi

"$ROOT"/converge apply --local -p "lang=japanese" "$SOURCE"

if [ ! -f greeting.txt ]; then
    echo "greeting.txt doesn't exist!"
    exit 1
fi

if [[ "$(cat greeting.txt)" != "もしもし" ]]; then
    echo "greeting.txt doesn't have the right content"
    echo "has '$(at greeting.txt)', want 'もしもし'"
    exit 1
fi

"$ROOT"/converge apply --local -p "lang=esperanto" "$SOURCE"

if [ ! -f greeting.txt ]; then
    echo "greeting.txt doesn't exist!"
    exit 1
fi

if [[ "$(cat greeting.txt)" != "hello" ]]; then
    echo "greeting.txt doesn't have the right content"
    echo "has '$(at greeting.txt)', want 'hello'"
    exit 1
fi

popd
