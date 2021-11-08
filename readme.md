## Setting development environment

### Prerequisites 

mantil-io Github account, and configured Github [access](https://docs.github.com/en/authentication/connecting-to-github-with-ssh).

Configured [AWS cli access](https://docs.aws.amazon.com/cli/latest/userguide/cli-chap-configure.html) to AWS account atoz.technology (477361877445).

### Repo and toolset

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

### Build

To build mantil cli from the repo root run:
``` shell
scripts/build-cli.sh
```

Binary will be located into `$(go env GOPATH)/bin` add that to jour PATH. Default Go location for GOPATH is ~/home/go so binary will be in ~/home/go/bin. Add something like `export PATH=$PATH:$(go env GOPATH)/bin` to you shell config.


To build mantil cli and deploy node functions run:

``` shell
scripts/deploy.sh
```

