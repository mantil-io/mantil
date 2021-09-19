#!/usr/bin/env bash -e

GIT_ROOT=$(git rev-parse --show-toplevel)
PARENT_DIR=$(cd "$GIT_ROOT/.."; pwd)

ASSETS_DIR=$GIT_DIR/internal/backend/assets

tf_module() {
    zip -j $1.zip $PARENT_DIR/terraform-aws-modules/$1/*.tf
    mv $1.zip $ASSETS_DIR/terraform/modules
}

echo "Building terraform modules..."
(cd $PARENT_DIR/terraform-aws-modules && git pull)

mkdir -p $ASSETS_DIR/terraform/modules
tf_module funcs
tf_module ws
tf_module backend-funcs
tf_module backend-iam

(cd $ASSETS_DIR && go-bindata -pkg=assets -fs terraform/modules/ terraform/templates/)

deploy_function() {
    echo "Deploying function $1..."
    env GOOS=linux GOARCH=amd64 go build -o bootstrap
    zip -j -y $1.zip bootstrap

    aws s3 cp $1.zip s3://mantil-downloads/functions/
    rm $1.zip
}

(cd $GIT_ROOT && git pull)

for d in $GIT_ROOT/functions/*; do
    func_name=$(basename $d)
    (cd $d && deploy_function $func_name)
done
