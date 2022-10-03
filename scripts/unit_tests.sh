#!/usr/bin/env sh

GIT_ROOT=$(git rev-parse --show-toplevel)
cd $GIT_ROOT

go test -v  ./cli/... ./domain/... ./kit/... ./node/...
