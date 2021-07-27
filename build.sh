#!/usr/bin/env bash -e

WORK_DIR=~/work
ASSETS_DIR=$WORK_DIR/mantil-backend/internal/assets

zip -j funcs.zip $WORK_DIR/terraform-aws-modules/funcs/*.tf
mv funcs.zip $ASSETS_DIR/terraform/modules

zip -j dynamodb.zip $WORK_DIR/terraform-aws-modules/dynamodb/*.tf
mv dynamodb.zip $ASSETS_DIR/terraform/modules

(cd $ASSETS_DIR && go-bindata -pkg=assets -fs terraform/modules/ terraform/templates/ aws/)
