#!/usr/bin/env bash -e

GIT_ROOT=$(git rev-parse --show-toplevel)
#$GIT_ROOT/scripts/deploy.sh --only-cli


function show_terminal() {
    C='\033[1;36m'
    NC='\033[0m'
    printf "${C}"
    echo "$@"
    printf "${NC}"
    $@
    echo
}

function show() {
    printf "## "
    echo "$@"
    echo "\`\`\`"
    $@
    echo "\`\`\`"
    echo
}


show mantil --version
show mantil --help

show mantil aws --help
show mantil aws install --help
show mantil aws uninstall --help

show mantil stage --help
show mantil stage new --help
show mantil stage destory --help

# project commands
show mantil new --help
show mantil deploy --help
show mantil env --help
show mantil invoke --help
show mantil logs --help
show mantil test --help
show mantil watch --help

# generate
show mantil generate --help
show mantil generate api --help
