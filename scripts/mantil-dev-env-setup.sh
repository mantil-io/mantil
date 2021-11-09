#!/usr/bin/env bash

git clone git@github.com:mantil-io/mantil.git
git clone --recurse-submodules git@github.com:mantil-io/team.mantil.com.git
git clone git@github.com:mantil-io/mantil.go.git
git clone git@github.com:mantil-io/template-excuses.git
git clone git@github.com:mantil-io/go-mantil-template.git

set -euo pipefail

cd mantil/scripts
brew bundle

cd ../../team.mantil.com/scripts
brew bundle
cd ..
npm install
