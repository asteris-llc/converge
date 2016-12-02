#!/usr/bin/env bash
set -eo pipefail

LENGTH=${LENGTH:-60}

# SEED CORPUS
#
# when modifying the fuzzer or the code, it might help to remove the existing
# corpus and start fresh. These are the examples for that fresh start.
test -d corpus || mkdir corpus
printf abc       > corpus/abc
printf abc123xyz > corpus/alphanum
printf 8080      > corpus/port
printf a-        > corpus/dash
printf a.        > corpus/dot
printf a_        > corpus/underscore

# non-latin unicode letters
printf ڛ > corpus/arabic
printf もしもし > corpus/kanji


# BUILD FUZZER
echo "-- building fuzzer --"
make names-fuzz.zip

echo "-- running fuzzer for $LENGTH seconds --"
go-fuzz -bin=./names-fuzz.zip -workdir=. &
PID=$!
sleep "$LENGTH"

echo "-- killing fuzz process --"
kill -2 $PID
sleep 0.25

echo 'fuzzing complete. Examine the output at your leisure.'
