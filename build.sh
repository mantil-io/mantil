#!/usr/bin/env bash -e

WORK_DIR=~/work
ASSETS_DIR=$WORK_DIR/mantil-backend/internal/assets

tf_module() {
    zip -j $1.zip $WORK_DIR/terraform-aws-modules/$1/*.tf
    mv $1.zip $ASSETS_DIR/terraform/modules
}

(cd $WORK_DIR/terraform-aws-modules && git pull)

tf_module funcs
tf_module dynamodb
tf_module backend-funcs
tf_module backend-iam

(cd $ASSETS_DIR && go-bindata -pkg=assets -fs terraform/modules/ terraform/templates/ aws/)
