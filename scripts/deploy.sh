#!/usr/bin/env bash
#
#
# About version.
# If on the exact tag version is git tag, functions are deployed to /functions/$tag path.
# If thare are commits after tag tag is apended with commit hash, functions are deployed to /dev/$tag.
# If repo is dirty username is added to the tag.
#
# Flags:
#   --only-cli                    just builds cli
#   --silent                      don't send slack notification

GIT_ROOT=$(git rev-parse --show-toplevel)

cd "$GIT_ROOT/cli"
# collect variables
tag=$(git describe --always)
# if we are exactly on the tag
on_tag=0; (git describe --exact-match > /dev/null 2>&1 && git diff --quiet) && { on_tag=1; }
mantil_bin="$(go env GOPATH)/bin/mantil"

tag=v0.1.24
on_tag=1

echo "> Building cli with tag=$tag dev=$USER on_tag=$on_tag"
go build -o "$mantil_bin" -ldflags "-X github.com/mantil-io/mantil/domain.tag=$tag -X github.com/mantil-io/mantil/domain.dev=$USER -X github.com/mantil-io/mantil/domain.ontag=$on_tag" -trimpath

if [[ $* == *--only-cli* ]]; then
   exit 0
fi

# set BUCKET, BUCKET2, RELEASE env variables
$mantil_bin --version # ensure that mantil is in the path
eval "$(MANTIL_ENV=1 $mantil_bin)"

if [ -n "$RELEASE" ]; then
   echo "> Releasing new cli version to homebrew"
   cd "$GIT_ROOT"
   (export tag=$tag dev=$USER on_tag=$on_tag; goreleaser release --rm-dist)
   scripts/copy_release_to_latest.sh $tag
fi

deploy_function() {
    env GOOS=linux GOARCH=arm64 go build -o bootstrap
    zip -j -y -q "$1.zip" bootstrap

    aws s3 cp --no-progress "$1.zip" "$BUCKET"
    if [ -n "$BUCKET2" ]; then
       aws s3 cp --no-progress "$1.zip" "$BUCKET2"
    fi
    rm "$1.zip"
}

echo "> Deploying functions to $BUCKET"
for d in $GIT_ROOT/node/functions/*; do
    func_name=$(basename $d)
    (cd $d && deploy_function $func_name)
done

if [[ $* == *--silent* ]]; then
   exit 0
fi

# slack notification for new published version
if [ -n "$RELEASE" ]; then
   curl -X POST -H 'Content-type: application/json' --data '{"text":"Mantil version '$tag' is released!"}' https://hooks.slack.com/services/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
fi
