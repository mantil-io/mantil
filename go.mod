module github.com/mantil-io/mantil

go 1.16

require (
	github.com/Microsoft/go-winio v0.4.17 // indirect
	github.com/ajg/form v1.5.1 // indirect
	github.com/alecthomas/jsonschema v0.0.0-20210920000243-787cd8204a0d
	github.com/aws/aws-lambda-go v1.27.0
	github.com/aws/aws-sdk-go-v2 v1.11.2
	github.com/aws/aws-sdk-go-v2/config v1.11.0
	github.com/aws/aws-sdk-go-v2/credentials v1.6.4
	github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue v1.4.4 // indirect
	github.com/aws/aws-sdk-go-v2/service/apigateway v1.8.0
	github.com/aws/aws-sdk-go-v2/service/apigatewaymanagementapi v1.4.0
	github.com/aws/aws-sdk-go-v2/service/cloudformation v1.10.1
	github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs v1.5.2
	github.com/aws/aws-sdk-go-v2/service/dynamodb v1.10.0
	github.com/aws/aws-sdk-go-v2/service/ec2 v1.19.0
	github.com/aws/aws-sdk-go-v2/service/iam v1.11.0
	github.com/aws/aws-sdk-go-v2/service/lambda v1.14.1
	github.com/aws/aws-sdk-go-v2/service/resourcegroupstaggingapi v1.5.1
	github.com/aws/aws-sdk-go-v2/service/s3 v1.21.0
	github.com/aws/aws-sdk-go-v2/service/ses v1.9.0
	github.com/aws/aws-sdk-go-v2/service/sts v1.11.1
	github.com/aws/smithy-go v1.9.0
	github.com/denisbrodbeck/machineid v1.0.1
	github.com/fasthttp-contrib/websocket v0.0.0-20160511215533-1f3b11f56072 // indirect
	github.com/fatih/color v1.12.0
	github.com/fatih/structs v1.1.0 // indirect
	github.com/gavv/httpexpect v2.0.0+incompatible
	github.com/go-git/go-git/v5 v5.4.2
	github.com/google/go-querystring v1.1.0 // indirect
	github.com/google/uuid v1.3.0
	github.com/gorilla/websocket v1.4.2 // indirect
	github.com/imkira/go-interpol v1.1.0 // indirect
	github.com/json-iterator/go v1.1.12
	github.com/kataras/jwt v0.1.2
	github.com/klauspost/compress v1.13.6 // indirect
	github.com/manifoldco/promptui v0.8.0
	github.com/mantil-io/mantil.go v0.1.9
	github.com/mattn/go-colorable v0.1.11
	github.com/mitchellh/mapstructure v1.4.3 // indirect
	github.com/moul/http2curl v1.0.0 // indirect
	github.com/nats-io/jsm.go v0.0.26
	github.com/nats-io/nats.go v1.13.0
	github.com/nats-io/nkeys v0.3.0
	github.com/olekukonko/tablewriter v0.0.5
	github.com/onsi/ginkgo v1.16.5 // indirect
	github.com/onsi/gomega v1.17.0 // indirect
	github.com/pkg/errors v0.9.1
	github.com/qri-io/jsonschema v0.2.2-0.20210831022256-780655b2ba0e
	github.com/radovskyb/watcher v1.0.7
	github.com/sergi/go-diff v1.2.0
	github.com/spf13/cobra v1.2.1
	github.com/stretchr/testify v1.7.0
	github.com/valyala/fasthttp v1.31.0 // indirect
	github.com/xeipuuv/gojsonschema v1.2.0 // indirect
	github.com/yalp/jsonpath v0.0.0-20180802001716-5cc68e5049a0 // indirect
	github.com/yudai/gojsondiff v1.0.0 // indirect
	github.com/yudai/golcs v0.0.0-20170316035057-ecda9a501e82 // indirect
	github.com/yudai/pp v2.0.1+incompatible // indirect
	golang.org/x/crypto v0.0.0-20211202192323-5770296d904e // indirect
	golang.org/x/mod v0.4.2
	golang.org/x/sys v0.0.0-20211109184856-51b60fd695b3 // indirect
	golang.org/x/term v0.0.0-20210927222741-03fcf44c2211
	golang.org/x/time v0.0.0-20200630173020-3af7569d3a1e // indirect
	gopkg.in/yaml.v2 v2.4.0
)

//replace github.com/mantil-io/mantil.go => ../mantil.go
