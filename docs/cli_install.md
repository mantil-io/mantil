**⚠️ Notice: This documentation is deprecated, please visit [docs.mantil.com](https://docs.mantil.com/cli_install) to get the latest version!**

# Mantil CLI Installation

Mantil consists of two main components, node and CLI. CLI is the Mantil binary you
install on your local machine while [the node](aws_install.md) is located in AWS. To install CLI, follow the steps below for your OS.

## Homebrew (Mac and Linux)

Use [Homebrew](https://brew.sh) to install the latest version:

```
brew tap mantil-io/mantil
brew install mantil
```

## Direct Download (Linux, Windows and Mac)

Below are the available downloads for the latest version of Mantil. Please
download the right package for your operating system and architecture.

Mantil is distributed as a single binary. Install Mantil by extracting it and
moving it to a directory included in your system's PATH.

| OS      | Architecture | Download link                                                                                                                |
| --------| ------------ | ---------------------------------------------------------------------------------------------------------------------------- |
| Darwin  | arm64        | [mantil_Darwin_arm64.tar.gz](https://s3.eu-central-1.amazonaws.com/releases.mantil.io/latest/mantil_Darwin_arm64.tar.gz)     |
| Darwin  | x86_64       | [mantil_Darwin_x86_64.tar.gz](https://s3.eu-central-1.amazonaws.com/releases.mantil.io/latest/mantil_Darwin_x86_64.tar.gz)   |
| Linux   | x86_64       | [mantil_Linux_x86_64.tar.gz](https://s3.eu-central-1.amazonaws.com/releases.mantil.io/latest/mantil_Linux_x86_64.tar.gz)     |
| Linux   | i386         | [mantil_Linux_i386.tar.gz](https://s3.eu-central-1.amazonaws.com/releases.mantil.io/latest/mantil_Linux_i386.tar.gz)         |
| Linux   | arm64        | [mantil_Linux_arm64.tar.gz](https://s3.eu-central-1.amazonaws.com/releases.mantil.io/latest/mantil_Linux_arm64.tar.gz)       |
| Linux   | armv6        | [mantil_Linux_arm.tar.gz](https://s3.eu-central-1.amazonaws.com/releases.mantil.io/latest/mantil_Linux_armv6.tar.gz)           |
| Windows | x86_64       | [mantil_Windows_x86_64.tar.gz](https://s3.eu-central-1.amazonaws.com/releases.mantil.io/latest/mantil_Windows_x86_64.tar.gz) |
| Windows | i386         | [mantil_Windows_i386.tar.gz](https://s3.eu-central-1.amazonaws.com/releases.mantil.io/latest/mantil_Windows_i386.tar.gz)     |


### An example for Linux x86_64

```
wget https://s3.eu-central-1.amazonaws.com/releases.mantil.io/latest/mantil_Linux_x86_64.tar.gz
tar xvfz mantil_Linux_x86_64.tar.gz
mv mantil /usr/local/bin
```


<p align="right"> <a href="https://github.com/mantil-io/mantil/tree/master/docs#mantil-documentation">↵ Back to Documentation Home!</a></p>
