#!/usr/bin/env bash -e
#
#
# Helper script for running tests
#
# Flags:
# --only-cli to run tests only for cli commands
# --only-api to run tests only for api functions

GIT_ROOT=$(git rev-parse --show-toplevel)

run_cli_tests() {
    echo "> Running CLI tests"
    cd "$GIT_ROOT/cli/cmd"
    go test -v .
}

run_api_tests() {
   echo "> Running API tests"
    for d in $GIT_ROOT/api/*; do
        echo "> Running tests for $(basename $d)"
        (cd $d && go test -v .)
    done
}

only_cli=1
only_api=1
no_flags=0

if [[ $* == *--only-cli* ]]; then only_cli=0; fi
if [[ $* == *--only-api* ]]; then only_api=0; fi
if [[ $only_cli -eq 0 || $only_api -eq 0 ]]; then no_flags=1; fi

if [[ $only_cli -eq 0 || $no_flags -eq 0 ]]; then
    run_cli_tests
fi

if [[ $only_api -eq 0 || $no_flags -eq 0 ]]; then
    run_api_tests
fi


