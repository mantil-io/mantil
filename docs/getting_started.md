## Setup

Beside Mantil cli you will need access to an AWS account. AWS credentials are
needed just for initial setting up Mantil in your account. After the initial
setup the other commands won't need your AWS credentials.


To install Mantil in a region of an AWS account use `mantil aws install`. You
can provide AWS credentials in three different ways:

- As command line arguments:

```
mantil aws install --aws-access-key-id=AKIAIOSFODNN7EXAMPLE \
                   --aws-secret-access-key=wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY \
                   --aws-region=us-east-1
```

- By using environment variables AWS_ACCESS_KEY_ID, AWS_SECRET_ACCESS_KEY and
  AWS_DEFAULT_REGION from current shell. Those variables should be set before
  calling:

```
mantil aws install --aws-env
```

- By using named profile from your AWS configuration:

```
mantil aws install --aws-profile=my-named-profile
```

This command will create: one S3 bucket, few Lambda functions, two API Gateways,
few IAM roles, one DynamoDB table, CloudWatch log groups for Lambda functions,
one CloudFormation stack. All resources will have prefix 'mantil-' in the name.
The opposite command `mantil aws uninstall` will clean-up all those resources
and leave AWS account in initial state.

After this we have Mantil installation in a region of an AWS account. That we
will call Mantil Node, or just Node.

## Your first Mantil project

Run `mantil new` command to create project structure on the local computer. It
creates [project structure](https://github.com/mantil-io/docs/blob/main/concepts.md#project) with demo ping
API which we will use later in this guide.

For example when I run `mantil new my-project` command in my ~/mantil folder
expected output is:

```
➜ mantil new my-project
Creating project my-project from template ping...
Cloning into ~/mantil/my-project and replacing import paths with my-project...
Project initialized in ~/mantil/my-project
```

All project commands are intended to be used from somewhere in the project tree.
So enter the project:

```
cd my-project
```

## Project Stage

One Mantil project can have multiple deployments, each called deployment stage.
So we can have stage for development, staging, production and so on. Each stage
requires some resources on AWS and each stage is completely isolated from all
other stages.

To create first stage, named development use:

```
mantil stage new development
```

This operation usually takes about 1-2 minute to complete.\
Upon completion we have fully functional demo API on the AWS infrastructure.
Demo is build from [this](https://github.com/mantil-io/template-ping)
template project.

Code of the ping API is located in api/ping/ping.go file.\
To get API entrypoint run:

```
mantil env --url
```

To execute
[Default](https://github.com/mantil-io/template-ping/blob/b7200b4663116e26edde4076bde4729b9cb3f077/api/ping/ping.go#L19)
method in ping API you can run:

```
➜ curl -X POST $(mantil env --url)/ping
pong%
```

or use `mantil invoke` for less typing and more features:

```
➜ mantil invoke ping
200 OK
pong
```

Celebrate! You have just created your first fully functional Mantil serverless
application.

## Exploring demo project

[Hello](https://github.com/mantil-io/template-ping/blob/b7200b4663116e26edde4076bde4729b9cb3f077/api/ping/ping.go#L26)
method is here to demonstrate calling method with some data:

```
➜ mantil invoke ping/hello --data "World"
200 OK
Hello, World
```

or if you prefer curl:

```
➜ curl -X POST $(mantil env --url)/ping/hello --data "World"
Hello, World%
```

[ReqRsp](https://github.com/mantil-io/template-ping/blob/b7200b4663116e26edde4076bde4729b9cb3f077/api/ping/ping.go#L41)
demonstrates JSON formatted request/response:

```
➜ mantil invoke ping/reqrsp --data '{"name":"World"}'
200 OK
{
   "Response": "Hello, World"
}
```

The
[logs](https://github.com/mantil-io/template-ping/blob/b7200b4663116e26edde4076bde4729b9cb3f077/api/ping/ping.go#L62)
method demonstrates display of function logs with invoke command. If your Lambda
function is using logging, the log lines are captured and shown before command
output:

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

This project comes with demo integration tests. Run them with:

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

Test are located in test directory. When run they are using current Stage to
make requests and explore results.

## Deployment

Lets first make some change into project to enable deployment. For example
change "pong" string in Default method (file: api/ping/ping.go) to something
else. I'll change it to "my-project" and than deploy changes with:

```
➜ mantil deploy
==> Building...
ping

==> Uploading...
ping

==> Updating...
ping

Build time: 1.206s, upload: 1.214s (6.2 MiB), update: 1.82s
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

Generating functions/second/main.go...
Generating functions/second/.gitignore...
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


