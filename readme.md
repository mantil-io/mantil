<img src="https://github.com/mantil-io/mantil/blob/master/docs/images/mantil-logo-lockup-1-orange-RGB.png" height="120px">
<p>
Serverless development kit for Go and AWS Lambda
</p>

[![License][License-Image]][License-Url] [![Slack][Slack-Image]][Slack-Url] ![test](https://github.com/mantil-io/mantil/actions/workflows/test.yml/badge.svg)

[License-Url]: https://github.com/mantil-io/mantil/blob/master/LICENSE
[License-Image]: https://img.shields.io/badge/license-MIT-blue

[Slack-Image]: https://img.shields.io/badge/chat-on%20slack-green
[Slack-Url]: https://join.slack.com/t/mantilcommunity/shared_invite/zt-z3iy0lsn-7zD_6nqEucsgygTvHmnxAw


#

Cloud-native development demands a new approach to building, launching and
observing cloud apps. Mantil is a modern Go toolset for creating and managing
AWS Lambda projects.

In this early version, [Mantil](https://www.mantil.com) addresses fundamental issues often encountered
while building and launching the apps:
* setting up a new AWS Lambda project from scratch or an existing template
* setting up a local development environment and tying everything with AWS
* deploying the app on every change
* code testing via standard go tests or by invoking a specific function
* getting logs instantly
* supporting multiple development stages and parallel lines of deployment

Please, go and try it! [Let us know](mailto:support@mantil.com?subject=Mantil%20feedback) your thoughts.

# Installation

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
| Linux   | arm          | [mantil_Linux_arm.tar.gz](https://s3.eu-central-1.amazonaws.com/releases.mantil.io/latest/mantil_Linux_arm.tar.gz)           |
| Windows | x86_64       | [mantil_Windows_x86_64.tar.gz](https://s3.eu-central-1.amazonaws.com/releases.mantil.io/latest/mantil_Windows_x86_64.tar.gz) |
| Windows | i386         | [mantil_Windows_i386.tar.gz](https://s3.eu-central-1.amazonaws.com/releases.mantil.io/latest/mantil_Windows_i386.tar.gz)     |


### An example for Linux x86_64

```
wget https://s3.eu-central-1.amazonaws.com/releases.mantil.io/latest/mantil_Linux_x86_64.tar.gz
tar xvfz mantil_Linux_x86_64.tar.gz
mv mantil /usr/local/bin
```

# Documentation

We suggest to check out [Getting started](docs/getting_started.md) and familiarize yourself with [General
Concepts](docs/concepts.md).

* [Getting Started](docs/getting_started.md)
* [General Concepts](docs/concepts.md)
* [FAQ](docs/faq.md)
* [Other](docs/readme.md) documentation topics

Start exploring by creating Mantil project from one of the templates:
* [ping](https://github.com/mantil-io/template-ping) - default template for new Mantil projects
* [excuses](https://github.com/mantil-io/template-excuses) - UI and environment variables showcase
* [chat](https://github.com/mantil-io/template-chat) - demonstrates WebSocket Mantil API interface
* [todo](https://github.com/mantil-io/template-todo) - simple todo app showcasing persistent key/value storage
* [github-to-slack](https://github.com/mantil-io/template-github-to-slack) - example of serverless integration between GitHub and Slack

