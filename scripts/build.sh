#!/usr/bin/env bash -e

root=$(git rev-parse --show-toplevel)

cd "$root/cmd/mantil"
go build
cp mantil /usr/local/bin
