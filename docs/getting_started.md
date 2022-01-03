This guide has accompanying video available on [youtube](https://youtu.be/Fp64VgSLoTQ).

## Prerequisites

 * Go
 * Mantil [cli](https://github.com/mantil-io/mantil#installation)
 * AWS Account
 
We assume that you are Go programmer so you have Go installed. After that you
need to download Mantil cli and have access to an AWS account.

## Node setup

AWS credentials are needed just for initial setting up Mantil in your account.
After the initial setup the other commands won't need your AWS credentials.

To install Mantil in a region of an AWS account use `mantil aws install`. This
will create Mantil
[node](https://github.com/mantil-io/mantil/blob/master/docs/concepts.md#node) in
your AWS account. Node consists of a set of Lambda functions, API Gateway and a
S3 bucket. After the node is created all other communication is between cli and
the node. 

Mantil is not storing your AWS credentials they are only used during node
install and later uninstall. 

You can provide AWS credentials in three different ways:

- As command line arguments:

```
mantil aws install --aws-access-key-id=AKIAIOSFODNN7EXAMPLE \
                   --aws-secret-access-key=wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY \
                   --aws-region=us-east-1
```

- Set [environment
  variables](https://docs.aws.amazon.com/cli/latest/userguide/cli-configure-envvars.html)
  AWS_ACCESS_KEY_ID, AWS_SECRET_ACCESS_KEY, AWS_DEFAULT_REGION and instruct
  Mantil to use that environment:

```
export AWS_ACCESS_KEY_ID=AKIAIOSFODNN7EXAMPLE
export AWS_SECRET_ACCESS_KEY=wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY
export AWS_DEFAULT_REGION=us-east-1
mantil aws install --aws-env
```

- Allow Mantil to use a [named
  profile](https://docs.aws.amazon.com/cli/latest/userguide/cli-configure-profiles.html)
  from your AWS configuration (~/.aws/config):

```
mantil aws install --aws-profile=my-named-profile
```

Install action usually less than 2 minutes to complete.  
After install `mantil aws nodes` command will show that node: 

```
➜ mantil aws nodes
| NAME | AWS ACCOUNT  |  AWS REGION  |    ID    | VERSION |
|------|--------------|--------------|----------|---------|
| demo | 052548195718 | eu-central-1 | 7582352e | v0.2.5  |
```

A node is located in a region of an AWS account. You can have multiple nodes in
the same or different account.

To see what resources are created for the node run `mantil aws resources` command:
```
➜ mantil aws resources

Node demo
Resources:
|    NAME    |         TYPE         |     AWS RESOURCE NAME      |               CLOUDWATCH LOG GROUP               |
|------------|----------------------|----------------------------|--------------------------------------------------|
| setup      | Lambda Function      | mantil-setup-7582352e      | /aws/lambda/mantil-setup-7582352e                |
| authorizer | Lambda Function      | mantil-authorizer-7582352e | /aws/lambda/mantil-authorizer-7582352e           |
| deploy     | Lambda Function      | mantil-deploy-7582352e     | /aws/lambda/mantil-deploy-7582352e               |
| destroy    | Lambda Function      | mantil-destroy-7582352e    | /aws/lambda/mantil-destroy-7582352e              |
| security   | Lambda Function      | mantil-security-7582352e   | /aws/lambda/mantil-security-7582352e             |
| setup      | CloudFormation Stack | mantil-setup-7582352e      |                                                  |
| http       | API Gateway          | mantil-http-7582352e       | /aws/vendedlogs/mantil-http-access-logs-7582352e |
|            | S3 Bucket            | mantil-7582352e            |                                                  |
Tags:
|       KEY        |         VALUE          |
|------------------|------------------------|
| MANTIL_KEY       | 7582352e               |
| MANTIL_WORKSPACE | LhvoKl2bQEib2UFhs7ypIA |
```

All node resources have prefix 'mantil-' in the name and a random suffix. Node
consist of few Lambda functions, CloudFormation stack, API Gateway, S3 bucket
and CloudWatch log groups.

Uninstall command `mantil aws uninstall` will clean-up all those resources and
leave AWS account in initial state.


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


