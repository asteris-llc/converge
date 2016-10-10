#!/bin/bash

distro=$1

case "$distro" in 
"centos" | "redhat" | "fedora")
    docker run --rm  -v $PWD:/converge ${distro}:latest /converge/converge apply --local /converge/converge-centos.hcl
    ;;
"debian" | "ubuntu")
    docker run --rm  -v $PWD:/converge ${distro}:latest  /converge/converge apply --local /converge/converge-deb-ubuntu.hcl
    ;;
*)
    echo "unsupported distribution: ${distro}"
    exit 1
    ;;
esac

