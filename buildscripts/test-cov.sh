#!/usr/bin/env bash

set -e

TOP_LEVEL=$(git rev-parse --show-toplevel 2>/dev/null || echo ".")
COVERAGE="$TOP_LEVEL/coverage"

mkdir -p "$COVERAGE"
rm "$COVERAGE"/* 2>/dev/null || :

for d in $(go list ./... | grep -v 'pkg/generated\|tests'); do
    #TODO - Include -race while creating the coverage profile.
    profile="$COVERAGE/unit-coverage-$(date +%s%N).txt"
    go test -coverprofile="$profile" -covermode=atomic $d
done
