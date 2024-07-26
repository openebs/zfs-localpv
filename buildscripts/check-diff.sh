#!/usr/bin/env bash

set -e

git --no-pager diff --exit-code  pkg/generated pkg/apis/ deploy/yamls
