#!/usr/bin/env bash
set -euo pipefail

version=${1:-}
if test -z "$version"; then
    echo "usage: $0 VERSION"
    exit 1
fi

pagesCheckout=$(mktemp -d)
finish() {
    rm -rf "$pagesCheckout"
}
trap finish EXIT

git clone --branch=gh-pages "$(git remote get-url origin)" "$pagesCheckout"
if ! test -d "$pagesCheckout/$version"; then
    mkdir "$pagesCheckout/$version"
else
    rm -rf "${pagesCheckout:?}/$version"
fi

hugo --baseUrl="http://converge.aster.is/$version" --destination="$pagesCheckout/$version"

pushd "$pagesCheckout" > /dev/null
git add --all
if git commit -m "publish from $(git rev-parse HEAD) as $version"; then
    git push origin gh-pages
else
    echo "no changes to publish"
fi
popd > /dev/null

rm -rf "$pagesCheckout"
