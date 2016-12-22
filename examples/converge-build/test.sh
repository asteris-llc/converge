#!/bin/bash

#tests setting up a converge build environment on different linux distributions

distro=$1

case "$distro" in 
"centos" | "fedora"| "debian" | "ubuntu")
    docker run --rm  -v $(pwd):/converge ${distro}:latest /converge/converge apply --local /converge/converge-linux.hcl
    ;;
*)
    echo "unsupported distribution: ${distro}"
    exit 1
    ;;
esac

