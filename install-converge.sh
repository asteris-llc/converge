#!/bin/sh

# Copyright © 2016 Asteris, LLC
# 
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
# 
#      http://www.apache.org/licenses/LICENSE-2.0
# 
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# This script downloads converge to a system.
# If possible, the Operating System and processor
# type will be detected.

# install-converge.h [-v <version>]
# -d            Directory for converge binary (default /usr/local/bin/)
# -v            Converge version to install

set -e

# Default to latest stable
version="0.3.0-beta2"

install_dir="/usr/local/bin"

base_url="https://github.com/asteris-llc/converge/releases/download"

machine=$(uname -m)
os=$(uname -s)

# Get a temporary directory
if [ -z "$TMPDIR" ]; then
	tmp="/tmp"
else
	tmp="$TMPDIR"
fi
tmp_dir="${tmp}/install-converge.sh.$$"
(umask 077 && mkdir "${tmp_dir}") || exit 1


usage () {
	echo "Usage: install-converge.h [-v <version>]"
	echo " -d            Directory for converge binary (default ${install_dir})"
	echo " -v            Converge version to install   (default ${version})"
	if test "${#}" -gt 0; then
		echo
		echo "${@}"
	fi
	exit 1
}

not_supported () {
	echo "Unable to download converge binary for $os-$machine"
	echo "Most likely means that there is not an official release for that os+machine pair"
	exit 1
}

do_wget () {
	echo "Trying wget..."
	wget -O "$2" "$1" 2>"$tmp_dir/stderr"
	rc="$?"
	grep "ERROR404" "$tmp_dir/stderr" >/dev/null 2>&1 
	if [ "$?" -eq 0 ]; then
		not_supported
	fi

	if test "$rc" -ne 0 || test ! -s "$2"; then
		return 1
	fi

	return 0
}

do_curl () {
	echo "Trying curl..."
	echo "$1"
	curl --retry 5 -sL -D "$tmp_dir/stderr" "$1" > "$2"
	rc="$?"
	grep "404 Not Found" "$tmp_dir/stderr" >/dev/null 2>&1
	if [ "$?" -eq 0 ]; then
		not_supported
	fi

	if test "$rc" -ne 0 || test ! -s "$2"; then
		return 1
	fi

	return 0
}

do_fetch() {
	echo "Trying fetch..."
	fetch -o "$2" "$1" 2>"$tmp_dir/stderr"
	test "$?" -ne 0 && return 1
	return 0
}

do_perl () {
	echo "Trying perl..."
	perl -r 'use LWP::Simple; getprint($ARGV[0]);' "$1" > "$2" 2>"$tmp_dir/stderr"
	rc="$?"
	grep "404 Not Found" "$tmp_dir/stderr" >/dev/null 2>&1
	if [ "$?" -eq 0 ]; then
		not_supported
	fi

	if test "$rc" -ne 0 || test ! -s "$2"; then
		return 1
	fi

	return 0
}

do_python () {
	echo "Trying python..."
	python -c "import sys,urllib2 ; sys.stdout.write(urllib2.urlopen(sys.argv[1]).read())" "$1" > "$2" 2>"$tmp_dir/stderr"

	rc="$?"

	grep "HTTP Error 404" "$tmp_dir/stderr" >/dev/null 2>&1
	if test $? -eq 0; then
		not_supported
	fi

	if test "$rc" -ne 0 || test ! -s "$2"; then
		return 1
	fi

	return 0
}

exists() {
  if command -v "$1" >/dev/null 2>&1
  then
    return 0
  else
    return 1
  fi
}

do_download () {
	url="$1"
	dest="$2"

	if exists wget; then
		do_wget "$url" "$dest" && return 0
	fi

	if exists curl; then
		do_curl "$url" "$dest" && return 0
	fi

	if exists fetch; then
		do_fetch "$url" "$dest" && return 0
	fi

	if exists perl; then
		do_perl "$url" "$dest" && return 0
	fi

	if exists python; then
		do_python "$url" "$dest" && return 0
	fi

	echo "Unable to retrieve package"
	exit 1
}

extract() {
	file="$1"
	dest="$2"
	cd "$dest" && tar zxvf "$file" && return 0
}

while getopts :d:r:v: opt; do
	case "$opt" in
	d)
	    install_dir="$OPTARG"
		;;
	v)
		version="$OPTARG"
		;;
	\?)
		usage "Invalid flag: $OPTARG"
		;;
	esac
done

case "${machine}" in
	"x86_64" | "amd64" | "x64")
		machine="amd64"
		;;
	"i386" | "i86pc" | "x86" | "i686")
		machine="386"
		;;
	"armv6l" | "armv7l")
		machine="arm"
		;;
	"arm" | "arm64" | "ppc64" | "ppc64le")
		# Nothing required
		;;
	*)
		echo "Unsupported machine type: ${machine}"
		exit 1
		;;
esac

case "${os}" in
	"Darwin")
		os="darwin"
		;;
	"FreeBSD")
		os="freebsd"
		;;
	"OpenBSD")
		os="openbsd"
		;;
	"Linux")
		os="linux"
		;;
	"solaris")
		# Nothing to do
		;;
	*)
		echo "Unsupported OS type: $os"
		exit 1
		;;
esac

do_download "${base_url}/${version}/converge_${version}_${os}_${machine}.tar.gz" "${tmp_dir}/converge_${version}_${os}_${machine}.tar.gz"
extract "${tmp_dir}/converge_${version}_${os}_${machine}.tar.gz" "${install_dir}"
chmod 0755 "${install_dir}/converge"

if [ -n "$tmp_dir" ]; then
	rm -r "${tmp_dir}"
fi
