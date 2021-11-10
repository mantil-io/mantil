#!/usr/bin/env bash -e

tag=$(git describe --always)
on_tag=0; (git describe --exact-match > /dev/null 2>&1 && git diff --quiet) && { on_tag=1; }
mantil_bin="$(go env GOPATH)/bin/mantil.exe"

env GOOS=windows go build -o "$mantil_bin" -ldflags "-X github.com/mantil-io/mantil/domain.tag=$tag -X github.com/mantil-io/mantil/domain.dev=$USER -X github.com/mantil-io/mantil/domain.ontag=$on_tag" -trimpath
