#!/usr/bin/env bash -e

GIT_ROOT=$(git rev-parse --show-toplevel)
$GIT_ROOT/scripts/deploy.sh --only-cli
