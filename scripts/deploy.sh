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
TF_SCRIPT="$GIT_ROOT/scripts/build_terraform_modules.sh"


cd "$GIT_ROOT/cli"
# collect variables
tag=$(git describe)
commit=$(git rev-parse --short HEAD)
dirty=""
version=$tag
functions_path="dev/$tag"
on_tag=0
# if we are exactly on the tag
(git describe --exact-match > /dev/null 2>&1 && git diff --quiet) && { functions_path="functions/$tag";on_tag=1; }
# if local copy is dirty
(git diff --quiet) || { dirty="$USER";version="$tag-$dirty";functions_path="dev/$version"; }
if [[ $* == *--use-old-functions-path* ]]; then
   functions_path="functions"
fi
echo "> Building cli version: $version"
go build -o "$GOPATH/bin/mantil" -ldflags "-X main.commit=$commit -X main.tag=$tag -X main.dirty=$dirty -X main.version=$version -X main.functionsPath=$functions_path"
if [ $on_tag -eq 1 ]; then
   echo "> Releasing new cli version to homebrew"
   cd "$GIT_ROOT"
   (export commit=$commit tag=$tag dirty=$dirty version=$version functionsPath=$functions_path; goreleaser release --rm-dist)
fi
if [[ $* == *--only-cli* ]]; then
   exit 0
fi

source "$TF_SCRIPT"

deploy_function() {
    env GOOS=linux GOARCH=amd64 go build -o bootstrap
    zip -j -y -q "$1.zip" bootstrap

    aws s3 cp "$1.zip" "s3://mantil-downloads/$functions_path/"
    if [ $on_tag -eq 1 ]; then
       aws s3 cp "$1.zip" "s3://mantil-downloads/functions/latest/"
    fi
    rm "$1.zip"
}

echo "> Deploying functions to /$functions_path"
#(cd $GIT_ROOT && git pull)
for d in $GIT_ROOT/functions/*; do
    func_name=$(basename $d)
    (cd $d && deploy_function $func_name)
done
