**⚠️ Notice: This documentation is deprecated, please visit [docs.mantil.com](https://docs.mantil.com) to get the latest version!**

# Mantil Documentation 
Cloud-native development demands a new approach to building, launching, and observing cloud apps. [Mantil](https://www.mantil.com) is a modern open-source framework for writing serverless apps in Go. It allows you to quickly create and deploy applications that use AWS Lambda over a command line interface.

## Get Started
* [Mantil CLI Installation](cli_install.md)
* [Detailed AWS Setup](aws_install.md)
    - [AWS Credentials](https://github.com/mantil-io/mantil/blob/master/docs/aws_install.md#aws-credentials)
    - [Node Installation](https://github.com/mantil-io/mantil/blob/master/docs/aws_install.md#mantil-node-setup)
    - [Created Resources](https://github.com/mantil-io/mantil/blob/master/docs/aws_install.md#created-resources)
    - [Supported AWS Regions](https://github.com/mantil-io/mantil/blob/master/docs/aws_install.md#supported-aws-regions)
    - [AWS Costs](https://github.com/mantil-io/mantil/blob/master/docs/aws_install.md#aws-costs)
    - [Uninstall](https://github.com/mantil-io/mantil/blob/master/docs/aws_install.md#uninstall)
* [Your First Mantil Project](getting_started.md)

## General Concepts
* [Node](https://github.com/mantil-io/mantil/blob/master/docs/concepts.md#node)
* [Stage](https://github.com/mantil-io/mantil/blob/master/docs/concepts.md#stage)
* [Project Structure](https://github.com/mantil-io/mantil/blob/master/docs/concepts.md#project)
* [Developing in The Cloud](cloud_development.md)


## Usage
* [CLI Commands](commands/README.md)
* [Testing](testing.md): Unit tests, Integration tests, End to end test
* [Data Storage](data_storage.md)
* [Using a Mantil API](api.md): Rest API and WebSocket

## API Configuration
* [Environment Variables](api_configuration.md)
* [Scheduled Execution (cron jobs)](https://github.com/mantil-io/mantil/blob/master/docs/api_configuration.md#scheduled-execution)
* [Private API's](https://github.com/mantil-io/mantil/blob/master/docs/api_configuration.md#private-apis)
* [Custom Domain Names](https://github.com/mantil-io/mantil/blob/master/docs/api_configuration.md#custom-domain-names) 

## Examples
* [Chat](https://github.com/mantil-io/template-chat) - demonstrates WebSocket Mantil API interface
* [Todo](https://github.com/mantil-io/template-todo) - showcasing persistent key/value storage
* [Signup](https://github.com/mantil-io/example-signup) - example of simple signup workflow
* [Excuses](https://github.com/mantil-io/template-excuses) - UI and environment variables showcase
* [Github to Slack](https://github.com/mantil-io/template-github-to-slack) - example of serverless integration between GitHub and Slack
* [HN alerts](https://github.com/mantil-io/example-hn-alerts) - example of scheduled lambda function
* [Mongo Atlas](https://github.com/mantil-io/example-mongo-atlas) - example of using Mantil with Mongo Atlas
* [Presigned s3 upload](https://github.com/mantil-io/template-presigned-s3-upload) - template showing upload of files to S3 bucket through presigned URLs
* [NGS chat](https://github.com/mantil-io/example-ngs-chat) - example of chat implemented with [NATS](https://github.com/nats-io)


## More
* [Support and Troubleshooting](troubleshooting.md)
* [Anonymous Analytics](analytics.md)
* [Credits](credits.md) to the open source we use
* [FAQ](faq.md)
