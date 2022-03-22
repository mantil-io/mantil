**⚠️ Notice: This documentation is deprecated, please visit [docs.mantil.com](https://docs.mantil.com/Usage/data_storage) to get the latest version!**

# Data Storage

A key consideration when developing applications is choosing the right data storage option. For traditional server-based applications, the most common option is using a relational database. For serverless applications there are some additional things that need to be considered, such as connection management and pricing models. Here we will attempt to give an overview of the current options for serverless data storage and their advantages and disadvantages.

## S3

S3 is Amazon's serverless key/value object storage service that offers very high data availability, security, performance and virtually infinite scalability. Some common use cases include:
- storing assets such as images and videos
- backups and archives of infrequently used data
- analytics
- static website hosting

It is well suited for serverless workflows since it's not connection-based and offers an HTTP based interface.

While S3 is a very cost-effective solution for the use cases mentioned above, it is primarily designed for throughput and cannot replace a traditional database when it comes to latency and complex querying requirements.

For more information about S3, please refer to the official [user guide](https://docs.aws.amazon.com/AmazonS3/latest/userguide/Welcome.html).

To integrate S3 into your Mantil project, you can import the [mantil.go](https://github.com/mantil-io/mantil.go) package and use the [S3Bucket](https://github.com/mantil-io/mantil.go/blob/845476e8b2dae9333158fab6a48c7779423841a9/s3.go#L47) function to create a bucket. This will ensure that the created S3 bucket follows the standard Mantil resource naming convention and that it will be cleaned up when the stage is destroyed.

## DynamoDB

Amazon DynamoDB is a fully managed, serverless, key-value NoSQL database designed to run high-performance applications at any scale.

Compared to other NoSQL solutions like MongoDB, DynamoDB is more suited for serverless workloads since it:
- is not connection-based and offers a HTTP interface instead. This means that you don't have to worry about managing connection pools and cleaning up dead connections.
- scales seamlessly, meaning you don't have to worry about operational concerns such as hardware provisioning, setup/configuration, throughput capacity planning, replication, software patching, or cluster scaling.
- has a pay-per-use pricing model.

However, depending on your specific use case there are some potential downsides to consider:
- less flexibility when querying data. If you can't anticipate how your data will be accessed in advance, DynamoDB might not be a good choice.
- it has a steep learning curve for modelling data, especially when coming from a SQL background.

For simple use cases, Mantil offers a [KV store](https://github.com/mantil-io/mantil.go/blob/845476e8b2dae9333158fab6a48c7779423841a9/kv.go#L32) implementation backed by DynamoDB. This is used in some of our templates, such as:
https://github.com/mantil-io/template-todo
https://github.com/mantil-io/template-chat

For more complex use cases you can create a DynamoDB table by importing [mantil.go](https://github.com/mantil-io/mantil.go) and using the [DynamodbTable](https://github.com/mantil-io/mantil.go/blob/845476e8b2dae9333158fab6a48c7779423841a9/dynamo.go#L49) function. This will ensure that the created table follows the standard Mantil resource naming convention and that it will be cleaned up when the stage is destroyed.

As a quick example, you can create a Mantil project with an API named `dynamo` using the following code:
```
package dynamo

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/mantil-io/mantil.go"
)

const (
	tableName  = "table"
	partition  = "partition"
	primaryKey = "pk"
	sortKey    = "sk"
)

type Dynamo struct {
	client            *dynamodb.Client
	tableResourceName string
}

func New() *Dynamo {
	d := &Dynamo{}
	c, _ := mantil.DynamodbTable(tableName, primaryKey, sortKey)
	d.client = c
	d.tableResourceName = mantil.Resource(tableName).Name
	return d
}

type Item struct {
	ID   string
	Name string
}

func (d *Dynamo) Put(ctx context.Context, i *Item) error {
	av, err := attributevalue.MarshalMap(i)
	if err != nil {
		return err
	}
	av[primaryKey] = &types.AttributeValueMemberS{Value: partition}
	av[sortKey] = &types.AttributeValueMemberS{Value: i.ID}
	input := &dynamodb.PutItemInput{
		TableName: aws.String(d.tableResourceName),
		Item:      av,
	}
	_, err = d.client.PutItem(context.TODO(), input)
	return err
}

func (d *Dynamo) Get(key string) (*Item, error) {
	input := &dynamodb.GetItemInput{
		Key: map[string]types.AttributeValue{
			primaryKey: &types.AttributeValueMemberS{Value: partition},
			sortKey:    &types.AttributeValueMemberS{Value: key},
		},
		TableName: aws.String(d.tableResourceName),
	}
	result, err := d.client.GetItem(context.TODO(), input)
	if err != nil {
		return nil, err
	}
	var item Item
	if err := attributevalue.UnmarshalMap(result.Item, &item); err != nil {
		return nil, err
	}
	return &item, nil
}
```
After deploying this you can for example run:
```
mantil invoke dynamo/put -d '{"id": "item1", "name": "name1"}'
```
and then
```
mantil invoke dynamo/get -d "item1"
```
which will return
```
{
   "ID": "item1",
   "Name": "name1"
}
```
For more complicated use cases and information about DynamoDB, please refer to the official [developer guide](https://docs.aws.amazon.com/amazondynamodb/latest/developerguide/Introduction.html) and the [DynamoDB SDK docs for Go](https://pkg.go.dev/github.com/aws/aws-sdk-go-v2/service/dynamodb).

## Relational Databases

Relational databases are often the first choice for developers. They make data modelling easier since most data fits nicely in a relational model. They also allow for more flexible data access patterns which means that you don't have to anticipate how your data will be accessed in advance. Scalability is also usually not an issue since sharding and read replicas can go a long way. There are also plenty of fully managed RDBMS solutions like Amazon's RDS and Aurora services.

However, when trying to fit relational databases into a serverless workflow you might run into some issues:
- they are usually connection-based, which can be an issue when using them with lambda functions as they can create a large number of connections during usage peaks. This will cause slowdowns since relational databases are usually designed to support a limited number of persistent connections. You can mitigate this by implementing a connection pooling solution or using a managed database proxy like [RDS Proxy](https://aws.amazon.com/rds/proxy/) which will handle connection management, security and reduce failover times.
- they don't offer a pay-per-use billing model, instead you will usually pay an hourly rate based on instance size

One option worth mentioning here is [Aurora Serverless](https://aws.amazon.com/rds/aurora/serverless/) which is Amazon's serverless offering for their MySQL and PostgreSQL-compatible database [Aurora](https://aws.amazon.com/rds/aurora/). It offers an HTTP-based [Data API](https://docs.aws.amazon.com/AmazonRDS/latest/AuroraUserGuide/data-api.html) which doesn't require a persistent connection. It also scales capacity automatically which can reduce costs if database usage is unpredictable. There is however a minimum number of capacity units that need to be provisioned at all times which means that even if the database is not used there will be a minimum hourly cost.

A preview for [Aurora Serverless v2](https://docs.aws.amazon.com/AmazonRDS/latest/AuroraUserGuide/aurora-serverless-2.html) is also currently available offering faster and more granular scaling. However, it is still missing some features such as PostgreSQL and Data API support.

## Other NoSQL Databases

While DynamoDB is probably the best fit for serverless applications in most cases, there are also other popular NoSQL options like MongoDB and Amazon's DocumentDB which is MongoDB-compatible. They usually have less of a learning curve than DynamoDB, but they share the same issues as relational databases, namely being connection-based and not offering a pay-per-use pricing model.

Connection management is even more of an issue here as there are no managed proxy solutions available like RDS Proxy. Taking MongoDB as an example, there are some [best practices](https://docs.atlas.mongodb.com/best-practices-connecting-from-aws-lambda/) you can follow when connecting from a lambda function. You could also implement your own proxy solution as described in [this article](https://www.webiny.com/blog/using-aws-lambda-to-create-a-mongodb-connection-proxy-2bb53c4a0af4).

Another option to keep an eye out for is MongoDB's new [serverless offering](https://www.mongodb.com/cloud/atlas/serverless) which is currently in preview. Along with seamless auto-scaling it offers a pay-per-use pricing model based on the number of operations performed on the database and data storage/transfer.

<p align="right"> <a href="https://github.com/mantil-io/mantil/tree/master/docs#mantil-documentation">↵ Back to Documentation Home!</a></p>
