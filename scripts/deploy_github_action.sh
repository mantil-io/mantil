#!/usr/bin/bash -e

GIT_ROOT=$(git rev-parse --show-toplevel)

cd "$GIT_ROOT/cli"
# collect variables
tag=$(git describe --always)
# if we are exactly on the tag
on_tag=0
bin_path=${GOPATH:-/home/runner/go} # github action doesn't have GOPATH set

echo "> Building cli with tag=$tag dev=$USER on_tag=$on_tag"
go build -o "$bin_path/bin/mantil" -ldflags "-X github.com/mantil-io/mantil/domain.tag=$tag -X github.com/mantil-io/mantil/domain.dev=$USER -X github.com/mantil-io/mantil/domain.ontag=$on_tag" -trimpath
# set BUCKET, BUCKET2, RELEASE env variables
eval $(MANTIL_ENV=1 mantil)

deploy_function() {
    env GOOS=linux GOARCH=arm64 go build -o bootstrap
    zip -j -y -q "$1.zip" bootstrap
    aws s3 cp --no-progress "$1.zip" "$BUCKET"
    rm "$1.zip"
}

echo "> Deploying functions to $BUCKET"
for d in $GIT_ROOT/node/functions/*; do
    func_name=$(basename $d)
    (cd $d && deploy_function $func_name)
done

