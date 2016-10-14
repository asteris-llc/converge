#!/bin/bash

#tests setting up a converge build environment on different linux distributions

distro=$1

case "$distro" in 
"centos" | "fedora")
    docker run --rm  -v $(pwd):/converge ${distro}:latest /converge/converge apply --local /converge/converge-centos.hcl
    ;;
"debian" | "ubuntu")
    docker run --rm  -v $(pwd):/converge ${distro}:latest  /converge/converge apply --local /converge/converge-deb-ubuntu.hcl
    ;;
*)
    echo "unsupported distribution: ${distro}"
    exit 1
    ;;
esac

