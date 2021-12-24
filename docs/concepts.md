# Concepts

## Node

Before you start with the project you need to setup Mantil node on AWS. Node is
set of Lambda functions and supporting resources which will be used by Mantil to
deploy/upgrade/remove project Lambda functions and other project resources. Node
is installed in the specific region of an AWS Account.

## Project

Project is set of files in version controll system.  
This is example Mantil project structure with two API's: ping and second.

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

Project API is Go package located in _/api/[api-name]_ folder.
In this folder Mantil expects to find function _New()_ which has no parameters and has only one return value - struct or pointer to struct which represents API implementation structure.

For example this [Ping struct](https://github.com/mantil-io/template-ping/blob/master/api/ping/ping.go#L9) path is _/api/ping/ping.go_.

That struct is plain Go code without any dependencies to Mantil or AWS. All public methods of that structure are exposed as endpoint URL-s. Exported methods can have any of this signatures:

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

For reference see AWS
[Go handler](https://docs.aws.amazon.com/lambda/latest/dg/golang-handler.html)
documentation.

### Public folder

Public folder is place for you static site content. Put an _index.html_ file in
the _/public_ folder and it will be visible at the endpoint root.

### Build folder

Build folder is automatically generated on each deploy and contains main package for each API which is code that transforms your API's to Lambda
functions. Binaries of your functions named _bootstrap_ are also placed in this folder.
If you're using source control this folder should be untracked by adding it to your _.gitignore_ file since all data is generated on each deploy. This is automatically done for you when initializing project with _new_ command.

Code in build folder uses [Mantil Go
SDK](https://github.com/mantil-io/mantil.go) to transform you API's to AWS
Lambda functions.

### Config folder

In config folder _environment.yml_ is place where you can set environment
variables for each Stage. So you can configure different behavior in different
stages.  
_state.yml_ is project database file maintained by Mantil. It is stored in project
so you can history of all changes. You should not edit this file.

### Test

Test folder if where your API end to end tests are stored. Explore [ping](https://github.com/mantil-io/template-ping/blob/master/test/ping_test.go) example
to get idea how to create requests and explore results.


## Stage

Stage is actual installation of the project in AWS. A project can have
multiple stages. A stage for each developer, integration stage, production...  
Stage is placed on a node. Different stages in the same project can be placed on
different nodes.

## Endpoint

Endpoint is stage entrypoint from Internet. Each stage has two endpoints. One for
REST and the other for WebSocket communication
