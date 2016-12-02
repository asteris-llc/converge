#!/usr/bin/env bash
set -eo pipefail

LENGTH=${1:-60}

# POPULATE CORPUS
test -d corpus || mkdir corpus
printf "a=b"  > corpus/kv
printf "="    > corpus/singleequals
printf "a==b" > corpus/doubleequals

# BUILD FUZZER
echo "-- building fuzzer --"
make params-fuzz.zip

echo "-- running fuzzer for $LENGTH seconds --"
go-fuzz -bin=./params-fuzz.zip -workdir=. &
PID=$!
sleep "$LENGTH"

echo "-- killing fuzz process --"
kill -2 $PID
sleep 0.25

echo 'fuzzing complete. Examine the output at your leisure.'
