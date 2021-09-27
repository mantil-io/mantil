#1/usr/bin/env bash -e
#
#
# Build terraform modules used by API functions.
# Functions access modules through assets folder using go-bindata.

GIT_ROOT=$(git rev-parse --show-toplevel)
PARENT_DIR=$(cd "$GIT_ROOT/.."; pwd)
ASSETS_DIR=$GIT_ROOT/assets


echo "> Building terraform modules"
tf_module() {
    zip -j -q $1.zip $PARENT_DIR/terraform-aws-modules/$1/*.tf
    mv $1.zip $ASSETS_DIR/terraform/modules
}
(cd $PARENT_DIR/terraform-aws-modules && git pull)

mkdir -p $ASSETS_DIR/terraform/modules
tf_module funcs
tf_module backend-funcs
tf_module backend-iam
tf_module api

(cd $ASSETS_DIR && go-bindata -pkg=assets -fs terraform/modules/ terraform/templates/)
