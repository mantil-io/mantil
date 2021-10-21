package dto

import (
	"time"
)

type DeployRequest struct {
	AccountBucket      string
	FunctionsForUpdate []Function
	StageTemplate      *StageTemplate
}

type StageTemplate struct {
	Project                string
	Stage                  string
	Region                 string
	Bucket                 string
	BucketPrefix           string
	Functions              []Function
	AccountFunctionsBucket string
	AccountFunctionsPath   string
	ResourceSuffix         string
	ResourceTags           map[string]string
}

type Function struct {
	Name       string
	LambdaName string
	S3Key      string
	Runtime    string
	Handler    string
	MemorySize int
	Timeout    int
	Env        map[string]string
}

type DeployResponse struct {
	Rest         string
	Ws           string
	PublicBucket string
}

type DestroyRequest struct {
	Bucket       string
	Region       string
	ProjectName  string
	StageName    string
	BucketPrefix string
	ResourceTags map[string]string
}

const (
	RequestQueryParam = "r"
)

type SecurityRequest struct {
	CliRole         string
	Buckets         []string
	LogGroupsPrefix string
}

// credentials for aws sdk endpointcreds integration on the CLI
// fields are predefined by the SDK and can't be changed
// https://pkg.go.dev/github.com/aws/aws-sdk-go-v2/credentials/endpointcreds
type SecurityResponse struct {
	AccessKeyId     string
	SecretAccessKey string
	Token           string
	Expiration      *time.Time
}

type SetupRequest struct {
	Version         string
	Bucket          string
	FunctionsBucket string
	FunctionsPath   string
	PublicKey       string
	ResourceSuffix  string
	ResourceTags    map[string]string
	Destroy         bool
}

type SetupDestroyRequest struct {
	Bucket string
}

type SetupResponse struct {
	APIGatewayRestURL string
	APIGatewayWsURL   string
	CliRole           string
}
