**⚠️ Notice: This documentation is deprecated, please visit [docs.mantil.com](https://docs.mantil.com/aws_detailed_setup/aws_credentials) to get the latest version!**

# Detailed AWS Setup

Mantil consists of two main components, node and CLI. [CLI](cli_install.md) is the Mantil binary you
install on your local machine while the node is located in AWS. A node consists of Lambda functions and other AWS resources used for managing Mantil projects in the cloud and will be explained in detail in the upcoming chapter.

## AWS Credentials
You need to bring your own AWS account to work with Mantil. However, if you don't have one, you can easily create it by following [AWS instructions](https://portal.aws.amazon.com/billing/signup#/start). To install a node into your AWS account Mantil requires account credentials with IAM `AdministratorAccess` privileges. 

Mantil will never store your credentials; those are only used to set up a node into an AWS account. After the node is installed, all other communication is between the Mantil command line and the node. That means that node install/uninstall phases are the only time you need to provide AWS account credentials. 

Node functions have only necessary IAM permissions. All the resources created for the Mantil node (API Gateway, Lambda function, IAM roles) have the 'mantil-' prefix. You can list node resources by the `mantil aws resources` command.

<p align="right"> <a href="https://github.com/mantil-io/mantil/tree/master/docs#mantil-documentation">↵ Back to Documentation Home!</a></p>

#

## Mantil Node Setup
AWS credentials are needed just for initial setting up Mantil in your account.
After the initial setup, the other commands won't need your AWS credentials.

To install Mantil in a region of an AWS account, use `mantil aws install`. This
will create Mantil
[node](https://github.com/mantil-io/mantil/blob/master/docs/concepts.md#node) in
your AWS account. A node consists of a set of Lambda functions, API Gateway and a
S3 bucket. After the node is created, all other communication is between the CLI and
the node. 

Mantil is not storing your AWS credentials; they are only used during node
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

A node is located in a region of an AWS account. You can have multiple nodes in the same or different account.

<p align="right"> <a href="https://github.com/mantil-io/mantil/tree/master/docs#mantil-documentation">↵ Back to Documentation Home!</a></p>

#

## Created Resources
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

All node resources have prefix 'mantil-' in the name and a random suffix. 
Node consist of few Lambda functions, CloudFormation stack, API Gateway, S3 bucket and CloudWatch log groups.

<p align="right"> <a href="https://github.com/mantil-io/mantil/tree/master/docs#mantil-documentation">↵ Back to Documentation Home!</a></p>

#

## Supported AWS Regions
Mantil is using [Graviton (ARM) powered](https://aws.amazon.com/blogs/aws/aws-lambda-functions-powered-by-aws-graviton2-processor-run-your-functions-on-arm-and-get-up-to-34-better-price-performance/) Lambda functions. Mantil is available in every region where Graviton Lambda functions are [supported](https://github.com/mantil-io/mantil/blob/eafd1a09bade875e225b5f271cdb17f9211a970a/cli/controller/setup.go#L30):

> US East (N. Virginia), US East (Ohio), US West (Oregon), Europe (Frankfurt), Europe (Ireland), EU (London), Asia Pacific (Mumbai), Asia Pacific (Singapore), Asia Pacific (Sydney), Asia Pacific (Tokyo).

<p align="right"> <a href="https://github.com/mantil-io/mantil/tree/master/docs#mantil-documentation">↵ Back to Documentation Home!</a></p>

#

## AWS Costs
For trying Mantil, you can for sure stay within [free tier](https://aws.amazon.com/free/?all-free-tier.sort-by=item.additionalFields.SortRank&all-free-tier.sort-order=asc&awsf.Free%20Tier%20Types=*all&awsf.Free%20Tier%20Categories=*all) of all the AWS services. You have pretty generous monthly limits for many services when you create a new AWS account. The two most important you will use with Mantil are Lambda functions and API Gateway. Here are their free tier monthly limits:

> The Amazon API Gateway free tier includes one million API calls received for REST APIs, one million API calls received for HTTP APIs, and one million messages and 750,000 connection minutes for WebSocket APIs per month for up to 12 months.

> The AWS Lambda free tier includes one million free requests per month and 400,000 GB-seconds of compute time per month.

Until you don't have some significant user base or are not mining bitcoins in your Lambda function, you will for sure stay within the limits of the free tier. So trying Mantil will cost you nothing. 

<p align="right"> <a href="https://github.com/mantil-io/mantil/tree/master/docs#mantil-documentation">↵ Back to Documentation Home!</a></p>

#

## Uninstall
Uninstall command `mantil aws uninstall` will clean up all created resources and leave the AWS account in the initial state.
At this step, you will need to provide your AWS credentials again. There are three possible ways to do so that are already described in the node setup. 

After the uninstall, your account is in its original state. Mantil will remove anything it created.

<p align="right"> <a href="https://github.com/mantil-io/mantil/tree/master/docs#mantil-documentation">↵ Back to Documentation Home!</a></p>



