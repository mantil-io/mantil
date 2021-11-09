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


#
# /bin/bash -c "$(curl -fsSL https://gist.githubusercontent.com/ianic/26954cbd1f38275f3056db69f6424815/raw/e635a9e55b2d20a987a5bcfbb12556f788cc22fa/dev-env-setup.sh)"
#
# ssh-keygen -t ed25519 -C "igor.anic@gmail.com"
# eval "$(ssh-agent -s)"

# cat > ~/.ssh/config <<EOF
# Host *
#   AddKeysToAgent yes
#   UseKeychain yes
#   IdentityFile ~/.ssh/id_ed25519
# EOF


# ssh-add -K ~/.ssh/id_ed25519

# pbcopy < ~/.ssh/id_ed25519.pub
# # dodati na github:
# # https://docs.github.com/en/authentication/connecting-to-github-with-ssh/adding-a-new-ssh-key-to-your-github-account

# ssh -T git@github.com
