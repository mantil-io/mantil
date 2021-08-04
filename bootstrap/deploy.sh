#!/usr/bin/env bash -e

../build.sh

env GOOS=linux GOARCH=amd64 go build -o bootstrap
zip -j -y bootstrap.zip bootstrap

aws s3 cp bootstrap.zip s3://mantil-downloads/functions/
rm bootstrap.zip
