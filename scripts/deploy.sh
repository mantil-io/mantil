#!/usr/bin/env bash -e

root=$(git rev-parse --show-toplevel)

cd "$root/cmd/mantil"
GOOS=linux GOARCH=amd64 go build -o mantil-amd
aws s3 cp mantil-amd s3://mantil-downloads/mantil

go build
aws s3 cp mantil s3://mantil-downloads/mantil-osx
