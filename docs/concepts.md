**⚠️ Notice: This documentation is deprecated, please visit [docs.mantil.com](https://docs.mantil.com/concepts/node) to get the latest version!**

# Concepts

Two probably new concepts in Mantil are node and stage. So let's first familiarize yourself with them.

## Node

Mantil consists of two main components, node and CLI. CLI is mantil binary you
[install](cli_install.md) on your local machine.
Node is located in AWS. Node is set of functions used for managing Mantil
projects. First, you use CLI to [install a node into your AWS account](aws_install.md). After that, CLI issues commands, and the node executes them in the cloud.

Node is installed into a region of an AWS account. You can have multiple nodes
in the same or different AWS accounts. When you are setting project stage you
choose a node for that stage.

<p align="right"> <a href="https://github.com/mantil-io/mantil/tree/master/docs#mantil-documentation">↵ Back to Documentation Home!</a></p>

#

<!--

To install node you are using `mantil aws install`

* aws credentials
* koji su resursi
* install/uninstall

## Node

Before you start with the project you need to setup Mantil node on AWS. Node is
set of Lambda functions and supporting resources which will be used by Mantil to
deploy/upgrade/remove project Lambda functions and other project resources. Node
is installed in the specific region of an AWS Account.
-->

## Stage

The stage is an installation of a project into the cloud. Before creating stage, the project is
just a set of files. Once you create a stage that builds Lambda functions from your
api's, API Gateway for that functions and other supporting resources. With that
project gets a live endpoint where you can execute your api's.

The stage is located on a node. While creating stage [`mantil stage
new`](docs/commands/mantil_stage_new.md) you specify node for the stage. You can
have multiple stages of the project. Each can be located on a different node.

While working in the project, there is a notion of default stage. One stage is always
default, so it is the target of all other project commands. So, for example, when you
deploy/test/watch logs..., you execute those commands on the default stage.
To see all the stages and see which one is default there is [`mantil stage
ls`](docs/commands/mantil_stage_list.md) command, and to change default [`mantil
stage use`](docs/commands/mantil_stage_use.md).

### Stage endpoint

When created each stage gets two endpoints: HTTP and WebSocket. Endpoint is URL
where stage API's are exposed. 

HTTP endpoint will be something like:
_https://lh5rfrc3gf.execute-api.eu-central-1.amazonaws.com_, and WebSocket:
__wss://lh5rfrc3gf.execute-api.eu-central-1.amazonaws.com_.

<p align="right"> <a href="https://github.com/mantil-io/mantil/tree/master/docs#mantil-documentation">↵ Back to Documentation Home!</a></p>

#
## Project Structure

The project is a set of files on the disk and, of course, in a version control system.
You are creating new project with [`mantil new`](docs/commands/mantil_new.md).
That creates project structure and nothing more. Until you create a project
stage project don't have anything but files on the disk.

In Mantil we choose to favor convention over configuration. A big part of that
convention is in the Mantil project structure. Let's look into an example Mantil
project with two API's: ping and second.

```
├── api
│   ├── ping
│   │   ├── ping.go
│   │   └── ping_test.go
│   └── second
│       └── second.go
├── build
│   └── functions
│       ├── ping
|       |   ├── main.go
|       |   └── bootstrap
│       └── second
|           ├── main.go
|           └── bootstrap
├── config
│   ├── environment.yml
│   └── state.yml
├── go.mod
├── go.sum
├── public
│   ├── api.js
│   ├── index.html
│   └── style.css
└── test
    ├── init.go
    ├── ping_test.go
    └── second_test.go
```


### API

API folder is a set of Go packages where each package after deployment will become
exposed on the endpoint URL.

The package needs to have exported __New_ method, which returns a pointer to the
struct. That struct will be exposed at _endpoint/package-name_ URL. Where
endpoint is stage endpoint URL.

The package named ping from the example will be exposed at _endpoint/ping_. All
exported packages methods will be exposed at _endpoint/ping/method-name_. If the
package has the method named _Default_, that method is mapped to the package root
_endpoint/ping_. For example if the stage HTTP endpoint is
_https://lh5rfrc3gf.execute-api.eu-central-1.amazonaws.com_ URLs of the package
ping methods will be:

| URL                                                                      | Go method in ping.Ping struct |
| ------------------------------------------------------------------------ | ----------------------------- |
| https://lh5rfrc3gf.execute-api.eu-central-1.amazonaws.com/ping           | Default                       |
| https://lh5rfrc3gf.execute-api.eu-central-1.amazonaws.com/ping/hello     | Hello                         |
| https://lh5rfrc3gf.execute-api.eu-central-1.amazonaws.com/ping/reqrsp    | ReqRsp                        |


The method needs to have a specific signature to be used by Mantil in API URLs. Method
needs to be exported and must follow these rules:

* may take between 0 and two arguments.
* if there are two arguments, the first argument must satisfy the "context.Context" interface.
* may return between 0 and two arguments.
* if there are two return values, the second argument must be an error.
* if there is one return value, it must be an error.

valid signatures are:
```
   func ()
   func () error
   func (TIn) error
   func () (TOut, error)
   func (context.Context) error
   func (context.Context, TIn) error
   func (context.Context) (TOut, error)
   func (context.Context, TIn) (TOut, error)
```

The same as rules as for default AWS [Go
handler](https://docs.aws.amazon.com/lambda/latest/dg/golang-handler.html) with
the added convention that each method in the API struct is exposed as a URL path.

### Public folder

A public folder is a place for your static site content. Put an _index.html_ file in
the _/public_ folder, and it will be visible at the endpoint root.

### Build folder

The build folder is automatically generated on each deploy and contains the main package
for each API which is code that transforms your API's to Lambda functions.
Binaries of your functions named _bootstrap_ are also placed in this folder. If
you're using source control, this folder should be untracked by adding it to your
_.gitignore_ file since all data is generated on each deploy. This is
automatically done for you when initializing project with _new_ command.

Code in build folder uses [Mantil Go
SDK](https://github.com/mantil-io/mantil.go) to transform your API's to AWS
Lambda functions.

### Config folder

In config folder _environment.yml_ is a place where you can set environment
variables for each Stage. So you can configure different behavior in different
stages.  
_state.yml_ is a project database file maintained by Mantil. It is stored in the project
so you can see the history of all changes. You should not edit this file.

### Test

Test folder if where your API end to end tests are stored. Explore
[ping](https://github.com/mantil-io/template-ping/blob/master/test/ping_test.go)
example to get an idea of how to create requests and explore results. Read more about
[testing](testing.md) in Mantil.

<p align="right"> <a href="https://github.com/mantil-io/mantil/tree/master/docs#mantil-documentation">↵ Back to Documentation Home!</a></p>



