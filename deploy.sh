#!/usr/bin/env bash -e

WORK_DIR=~/work
ASSETS_DIR=$WORK_DIR/mantil-backend/internal/assets

tf_module() {
    zip -j $1.zip $WORK_DIR/terraform-aws-modules/$1/*.tf
    mv $1.zip $ASSETS_DIR/terraform/modules
}

echo "Building terraform modules..."
(cd $WORK_DIR/terraform-aws-modules && git pull)

tf_module funcs
tf_module dynamodb
tf_module backend-funcs
tf_module backend-iam

(cd $ASSETS_DIR && go-bindata -pkg=assets -fs terraform/modules/ terraform/templates/ aws/)

deploy_function() {
    echo "Deploying function $1..."
    env GOOS=linux GOARCH=amd64 go build -o bootstrap
    zip -j -y $1.zip bootstrap

    aws s3 cp $1.zip s3://mantil-downloads/functions/
    rm $1.zip
}

(cd $WORK_DIR/mantil-backend && git pull)

for d in $WORK_DIR/mantil-backend/functions/*; do
    func_name=$(basename $d)
    (cd $d && deploy_function $func_name)
done

(cd $WORK_DIR/mantil-backend/setup && deploy_function setup)
