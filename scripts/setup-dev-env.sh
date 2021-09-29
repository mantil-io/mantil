#!/usr/bin/env bash
set -euo pipefail

# homebrew packages
root=$(git rev-parse --show-toplevel)
cd $root/scripts
brew bundle

