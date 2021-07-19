#!/usr/bin/env bash -e

WORK_DIR=~/work
ASSETS_DIR=$WORK_DIR/mantil-cli/internal/assets

zip -j funcs.zip $WORK_DIR/terraform-aws-modules/funcs/*.tf
mv funcs.zip $ASSETS_DIR/terraform/modules

zip -j dynamodb.zip $WORK_DIR/terraform-aws-modules/dynamodb/*.tf
mv dynamodb.zip $ASSETS_DIR/terraform/modules

(cd $ASSETS_DIR && go-bindata -pkg=assets -fs github/ terraform/modules/ terraform/templates/)

(cd $WORK_DIR/mantil-cli/cmd/mantil && go install)
