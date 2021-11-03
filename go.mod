module github.com/mantil-io/mantil

go 1.16

require (
	github.com/Microsoft/go-winio v0.4.17 // indirect
	github.com/alecthomas/jsonschema v0.0.0-20210920000243-787cd8204a0d
	github.com/aws/aws-lambda-go v1.27.0
	github.com/aws/aws-sdk-go-v2 v1.10.0
	github.com/aws/aws-sdk-go-v2/config v1.9.0
	github.com/aws/aws-sdk-go-v2/credentials v1.5.0
	github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue v1.3.0 // indirect
	github.com/aws/aws-sdk-go-v2/service/apigatewaymanagementapi v1.4.0
	github.com/aws/aws-sdk-go-v2/service/cloudformation v1.10.1
	github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs v1.5.2
	github.com/aws/aws-sdk-go-v2/service/dynamodb v1.6.0
	github.com/aws/aws-sdk-go-v2/service/ec2 v1.19.0
	github.com/aws/aws-sdk-go-v2/service/iam v1.10.1
	github.com/aws/aws-sdk-go-v2/service/lambda v1.10.0
	github.com/aws/aws-sdk-go-v2/service/resourcegroupstaggingapi v1.5.1
	github.com/aws/aws-sdk-go-v2/service/s3 v1.11.1
	github.com/aws/aws-sdk-go-v2/service/sts v1.8.0
	github.com/aws/smithy-go v1.8.1
	github.com/fatih/color v1.12.0
	github.com/go-git/go-git/v5 v5.4.2
	github.com/json-iterator/go v1.1.12
	github.com/kataras/jwt v0.1.2
	github.com/klauspost/compress v1.13.6 // indirect
	github.com/manifoldco/promptui v0.8.0
	github.com/mantil-io/mantil.go v0.0.0-20211103222953-486b32975598
	github.com/nats-io/jsm.go v0.0.26
	github.com/nats-io/nats.go v1.13.0
	github.com/nats-io/nkeys v0.3.0
	github.com/pkg/errors v0.9.1
	github.com/qri-io/jsonschema v0.2.2-0.20210831022256-780655b2ba0e
	github.com/radovskyb/watcher v1.0.7
	github.com/sergi/go-diff v1.2.0
	github.com/spf13/cobra v1.2.1
	github.com/stretchr/testify v1.7.0
	golang.org/x/crypto v0.0.0-20210921155107-089bfa567519 // indirect
	golang.org/x/mod v0.4.2
	golang.org/x/sys v0.0.0-20210630005230-0f9fa26af87c // indirect
	golang.org/x/term v0.0.0-20210615171337-6886f2dfbf5b // indirect
	golang.org/x/time v0.0.0-20200630173020-3af7569d3a1e // indirect
	gopkg.in/yaml.v2 v2.4.0
)

//replace github.com/mantil-io/mantil.go => ../mantil.go
