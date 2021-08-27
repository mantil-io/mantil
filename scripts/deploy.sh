#!/usr/bin/env bash -e

root=$(git rev-parse --show-toplevel)

cd "$root/cmd/mantil"
go build
aws s3 cp mantil s3://mantil-downloads/mantil-osx
