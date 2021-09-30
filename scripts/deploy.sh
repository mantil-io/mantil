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

GIT_ROOT=$(git rev-parse --show-toplevel)

cd "$GIT_ROOT/cli"
# collect variables
tag=$(git describe)
# if we are exactly on the tag
on_tag=0; (git describe --exact-match > /dev/null 2>&1 && git diff --quiet) && { on_tag=1; }

echo "> Building cli with tag=$tag dev=$USER on_tag=$on_tag"
go build -o "$GOPATH/bin/mantil" -ldflags "-X main.tag=$tag -X main.dev=$USER -X main.ontag=$on_tag"
# set BUCKET, BUCKET2, RELEASE env variables
eval $(MANTIL_ENV=1 mantil)

if [ -n "$RELEASE" ]; then
   echo "> Releasing new cli version to homebrew"
   cd "$GIT_ROOT"
   (export tag=$tag dev=$USER on_tag=$on_tag; goreleaser release --rm-dist)
fi
if [[ $* == *--only-cli* ]]; then
   exit 0
fi

deploy_function() {
    env GOOS=linux GOARCH=amd64 go build -o bootstrap
#    zip -j -y -q "$1.zip" bootstrap

#    aws s3 cp --no-progress "$1.zip" "$BUCKET"
    if [ -n "$BUCKET2" ]; then
       aws s3 cp --no-progress "$1.zip" "$BUCKET2"
    fi
#    rm "$1.zip"
}

echo "> Deploying functions to $BUCKET"
for d in $GIT_ROOT/functions/*; do
    func_name=$(basename $d)
    (cd $d && deploy_function $func_name)
done
