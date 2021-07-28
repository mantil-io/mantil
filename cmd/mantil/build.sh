#!/usr/bin/env bash -e

WORK_DIR=~/work
ASSETS_DIR=$WORK_DIR/mantil-cli/internal/assets

(cd $ASSETS_DIR && go-bindata -pkg=assets -fs)

(cd $WORK_DIR/mantil-cli/cmd/mantil && go install)
