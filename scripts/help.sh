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

show mantil new --help

show mantil deploy --help
show mantil destroy --help
show mantil env --help
show mantil invoke --help
show mantil logs --help
show mantil test --help
show mantil watch --help

show mantil generate --help
show mantil generate api --help
