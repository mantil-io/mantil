## Setting development environment

you have mantil-io Github accountu, and configured Github [access](https://docs.github.com/en/authentication/connecting-to-github-with-ssh).



``` shell
cd ~
mkdir mantil-io
cd mantil-io
git clone git@github.com:mantil-io/mantil.git
cd mantil
cd scripts
brew bundle
cd ..
```


To build mantil cli from the repo root run:
``` shell
scripts/build-cli.sh
```

Binary will be located into `$(go env GOPATH)/bin` add that to jour PATH. Default Go location for GOPATH is ~/home/go so binary will be in ~/home/go/bin.

