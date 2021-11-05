#!/usr/bin/env bash

GIT_ROOT=$(git rev-parse --show-toplevel)
cd "$GIT_ROOT"
MANTIL_GEN_DOC="$GIT_ROOT/../mantil-io.github.io" mantil
cd "$GIT_ROOT/../mantil-io.github.io"
