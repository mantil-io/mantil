**⚠️ Notice: This documentation is deprecated, please visit [docs.mantil.com](https://docs.mantil.com/getting_started) to get the latest version!**

This guide has accompanying video available on [youtube](https://youtu.be/Fp64VgSLoTQ).

## Prerequisites

 * Go
 * Mantil [cli](installation.md)
 * Mantil [node](aws_install.md)
 
We assume that you are Go programmer so you have Go installed. After that you
need to download Mantil cli and set up Mantil node on your AWS account. 

## Your first Mantil project

Create a new project with `mantil new` command. It just creates [project structure](https://github.com/mantil-io/docs/blob/main/concepts.md#project).

```
➜ mantil new my-project

Your project is ready in ./my-project

➜ tree my-project
my-project
├── api
│   └── ping
│       ├── ping.go
│       └── ping_test.go
├── config
│   ├── environment.yml
│   └── state.yml
├── go.mod
├── go.sum
└── test
    ├── init.go
    └── ping_test.go

4 directories, 8 files
```

API folder is most interesting. Each Go package in the API folder, after
deployment, becomes part of you applications API interface.


All other project commands are intended to be used from somewhere in the project
tree. So enter the project now `cd my-project`.  


## Project Stage

One Mantil project can have multiple deployments, each called deployment stage.
So we can have stage for development, staging, production and so on. Each stage
requires some resources on AWS and each stage is completely isolated from all
other stages.  
Stage is created on the specified node (if you have just one node than you don't
need to say on which node).

To create first stage, named development run:

```
➜ mantil stage new development

Creating stage development on node demo
...

Deploy successful!
Build time: 625ms, upload: 728ms (5.4 MiB), update: 28.573s

Stage development is ready!
Endpoint: https://lh5rfrc3gf.execute-api.eu-central-1.amazonaws.com
```

This operation usually takes less than a minute to complete. Upon completion we
have fully functional API on the AWS infrastructure. 

To see what resources is created run `mantil aws resources --stage development`.

Endpoint from the command output is the api URL.
To test that API exists run:

```
➜ mantil invoke ping
200 OK
pong
```

You can get the same result with any other tool:
```
➜ curl -X POST $(mantil env --url)/ping
pong%
```

`mantil env --url` returns API endpoint.


Each Go package in the api folder becomes route in the project URL. Package is
expected to have exported New method which returns struct pointer. All exported
methods of that struct will become accessible on endpoint/package/method URL. If
there is a method named Default it is accessible on the endpoint/package
(without method name) URL.

In our example package name is ping and we have Default method:
```Go
func (p *Ping) Default() string {
	return "pong"
}
```


## Exploring demo project

To execute non-default method we need to add method name to the path. Here is example of calling another method, named 
[Hello](https://github.com/mantil-io/template-ping/blob/11ff351b83ded21b93e6bdb0bd409273ef6075a6/api/ping/ping.go#L27):

```
➜ mantil invoke ping/hello --data "World"
200 OK
Hello, World
```

Hello method is again simple string in string out. 
```Go
func (p *Ping) Hello(ctx context.Context, name string) (string, error) {
	return "Hello, " + name, nil
}
```

You can also use curl for calling any method:

```
➜ curl -X POST $(mantil env --url)/ping/hello --data "World"
Hello, World%
```

[ReqRsp](https://github.com/mantil-io/template-ping/blob/11ff351b83ded21b93e6bdb0bd409273ef6075a6/api/ping/ping.go#L42)
demonstrates JSON formatted request/response:

```
➜ mantil invoke ping/reqrsp --data '{"name":"World"}'
200 OK
{
   "Response": "Hello, World"
}
```

The
[logs](https://github.com/mantil-io/template-ping/blob/11ff351b83ded21b93e6bdb0bd409273ef6075a6/api/ping/ping.go#L61)
method demonstrates display of function logs with invoke command. If your Lambda
function is logging, the log lines are captured and shown before command output:

```
➜ mantil invoke ping/logs --data '{"name":"Foo"}'
λ start Logs method
λ req.Name: 'Foo'
λ end
200 OK
{
   "Response": "Hello, Foo"
}
```

## Testing

This project comes with integration tests. Run them with:

```
➜ mantil test
=== RUN   TestPing
    printer.go:54: POST https://b2vhijcf2d.execute-api.eu-central-1.amazonaws.com/ping
    printer.go:54: POST https://b2vhijcf2d.execute-api.eu-central-1.amazonaws.com/ping/hello
    printer.go:54: POST https://b2vhijcf2d.execute-api.eu-central-1.amazonaws.com/ping/hello
    printer.go:54: POST https://b2vhijcf2d.execute-api.eu-central-1.amazonaws.com/ping/reqrsp
    printer.go:54: POST https://b2vhijcf2d.execute-api.eu-central-1.amazonaws.com/ping/reqrsp2
    printer.go:54: POST https://b2vhijcf2d.execute-api.eu-central-1.amazonaws.com/ping/non-existent-method
--- PASS: TestPing (0.62s)
PASS
ok  	my-project/test	0.902s
```

Test are located in test directory. When run they are using current stage to
make requests and explore results.

## Deployment

Lets first make some change into project to enable deployment. For example
change "pong" string in Default method (file: api/ping/ping.go) to something
else. I'll change it to "my-project" and than deploy changes with:

```
➜ mantil deploy

Building and deploying my-project to stage development
Building...
Uploading changes...
	ping
Updating infrastructure...

Deploy successful!
Build time: 636ms, upload: 789ms (5.4 MiB), update: 1.618s
```

To test new behavior run invoke again:

```
➜ mantil invoke ping
200 OK
my-project
```

Run also `mantil test` again, it is failing because of this change:

```
➜ mantil test
=== RUN   TestPing
    printer.go:54: POST https://b2vhijcf2d.execute-api.eu-central-1.amazonaws.com/ping
    reporter.go:23:
        	Error Trace:	reporter.go:23
        	            				chain.go:21
        	            				string.go:115
        	            				ping_test.go:17
        	Error:
        	            	expected string equal to:
        	            	 "pong"

        	            	but got:
        	            	 "my-project"
        	Test:       	TestPing
    printer.go:54: POST https://b2vhijcf2d.execute-api.eu-central-1.amazonaws.com/ping/hello
    printer.go:54: POST https://b2vhijcf2d.execute-api.eu-central-1.amazonaws.com/ping/hello
    printer.go:54: POST https://b2vhijcf2d.execute-api.eu-central-1.amazonaws.com/ping/reqrsp
    printer.go:54: POST https://b2vhijcf2d.execute-api.eu-central-1.amazonaws.com/ping/reqrsp2
    printer.go:54: POST https://b2vhijcf2d.execute-api.eu-central-1.amazonaws.com/ping/non-existent-method
--- FAIL: TestPing (0.58s)
FAIL
exit status 1
FAIL	my-project/test	0.867s
```

## Working

There is `mantil watch` command to support this change/deploy/invoke cycle. It
is monitoring project files for changes. On each change it deploys project and
can call a method or run tests. Run:

```
mantil watch --method ping
```

and then change return string of the Default method and save changes few times
to get the feeling.

To create new API run generate add implementation and deploy.

```
➜ mantil generate api second
Generating function second
test/init.go already exists
Generating test/second_test.go...
Generating api/second/second.go...
```

Now edit api/second/second.go add methods and deploy.

## Cleanup

To remove project Stage from you AWS account run:

```
mantil stage destroy development
```

and after that `mantil aws uninstall` with the same attributes as in the first
`aws install` step. And that's all. Your AWS account is the initial state.


