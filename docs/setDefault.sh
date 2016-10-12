#!/usr/bin/env bash
set -euo pipefail

version=${1:-}
if [[ -z "$version" ]]; then
    echo "usage: $0 VERSION"
    exit 1
fi

pagesCheckout=$(mktemp -d)
finish() {
    rm -rf "$pagesCheckout"
}
trap finish EXIT

git clone --branch=gh-pages "$(git remote get-url origin)" "$pagesCheckout"

pushd "$pagesCheckout" > /dev/null
echo "<html><head><script>window.location = '/$version/';</script></head><body>Redirecting to <a href='/$version/'>$version</a></body></html>" > index.html
git add --all
if git commit -m "set $version as default"; then
    git push origin gh-pages
else
    echo "no changes to publish"
fi
popd > /dev/null
