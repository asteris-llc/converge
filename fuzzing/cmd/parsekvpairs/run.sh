#!/usr/bin/env bash
# run from root directory

go-fuzz-build -func=ParseKVPairFuzz github.com/asteris-llc/converge/cmd
go-fuzz -bin=./cmd-fuzz.zip -workdir=fuzzing/cmd/parsekvpairs
