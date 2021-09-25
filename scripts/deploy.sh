#!/usr/bin/env bash -e
#
#
# About version.
# If on the exact tag version is git tag, functions are deployed to /functions/$tag path.
# If thare are commits after tag tag is apended with commit hash, functions are deployed to /dev/$tag.
# If repo is dirty username is added to the tag.
#
# Flags:
#   --only-cli                    just builds cli
#   --use-old-functions-path      deploys functions to the /functions

GIT_ROOT=$(git rev-parse --show-toplevel)
PARENT_DIR=$(cd "$GIT_ROOT/.."; pwd)
ASSETS_DIR=$GIT_ROOT/assets


cd "$GIT_ROOT/cmd/mantil"
# collect variables
tag=$(git describe)
commit=$(git rev-parse --short HEAD)
dirty=""
version=$tag
functions_path="dev/$tag"
# if we are exactly on the tag
(git describe --exact-match > /dev/null 2>&1 && git diff --quiet) && functions_path="functions/$tag"
# if local copy is dirty
(git diff --quiet) || { dirty="$USER";version="$tag-$dirty";functions_path="dev/$version"; }
if [[ $* == *--use-old-functions-path* ]]; then
   functions_path="functions"
fi
echo "> Building cli version: $version"
go build -o "$GOPATH/bin/mantil" -ldflags "-X main.commit=$commit -X main.tag=$tag -X main.dirty=$dirty -X main.version=$version -X main.functionsPath=$functions_path"
if [[ "$version" == "$tag" ]]; then
   echo "> Releasing new cli version to homebrew"
   cd "$GIT_ROOT"
   (export commit=$commit tag=$tag dirty=$dirty version=$version functionsPath=$functions_path; goreleaser release --rm-dist)
fi
if [[ $* == *--only-cli* ]]; then
   exit 0
fi

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

deploy_function() {
    env GOOS=linux GOARCH=amd64 go build -o bootstrap
    zip -j -y -q "$1.zip" bootstrap

    aws s3 cp "$1.zip" "s3://mantil-downloads/$functions_path/"
    rm "$1.zip"
}

echo "> Deploying functions to /$functions_path"
#(cd $GIT_ROOT && git pull)
for d in $GIT_ROOT/functions/*; do
    func_name=$(basename $d)
    (cd $d && deploy_function $func_name)
done
