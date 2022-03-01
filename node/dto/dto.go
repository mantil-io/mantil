package dto

import (
	"time"
)

type DeployRequest struct {
	ProjectName        string
	NodeBucket         string
	FunctionsForUpdate []Function
	StageTemplate      *StageTemplate
}

type StageTemplate struct {
	Project             string
	Stage               string
	Region              string
	Bucket              string
	BucketPrefix        string
	Functions           []Function
	NodeFunctionsBucket string
	NodeFunctionsPath   string
	ResourceSuffix      string
	ResourceTags        map[string]string
	WsEnv               map[string]string
	AuthEnv             map[string]string
	HasPublic           bool
	NamingTemplate      string
	PublicBucketName    string
	CustomDomain        CustomDomain
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
	Cron       string
	EnableAuth bool
}

type CustomDomains struct {
	Http CustomDomain
	Ws   CustomDomain
}

type CustomDomain struct {
	DomainName       string
	CertDomain       string
	HostedZoneDomain string
	HttpSubdomain    string
	WsSubdomain      string
}

type DeployResponse struct {
	Rest         string
	Ws           string
	PublicBucket string
}

type DestroyRequest struct {
	Bucket                string
	Region                string
	ProjectName           string
	StageName             string
	BucketPrefix          string
	ResourceTags          map[string]string
	CleanupBucketPrefixes []string
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
	Version            string
	BucketConfig       SetupBucketConfig
	FunctionsBucket    string
	FunctionsPath      string
	ResourceSuffix     string
	NamingTemplate     string
	APIGatewayLogsRole string
	AuthEnv            map[string]string
	ResourceTags       map[string]string
	GithubUser         string
	GithubOrg          string
}

type SetupBucketConfig struct {
	Name         string
	ExpirePrefix string
	ExpireDays   int
}

type SetupDestroyRequest struct {
	Bucket string
}

type SetupResponse struct {
	APIGatewayRestURL string
	CliRole           string
}
