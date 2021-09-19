#!/usr/bin/env bash
set -euo pipefail

# homebrew packages
root=$(git rev-parse --show-toplevel)
cd $root/scripts
brew bundle

# go-bindata
go install github.com/go-bindata/go-bindata/go-bindata@latest
which go-bindata >> /dev/null || (
    which $GOPATH/bin/go-bindata >>/dev/null && echo "go-bindata from $GOPATH/bin/go-bindata is not in PATH!" && exit 1
) || (
    which $HOME/go/bin/go-bindata >>/dev/null && echo "go-bindata from $HOME/go/go-bindata is not in PATH!" && exit 1
)
