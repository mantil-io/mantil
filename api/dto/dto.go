package dto

import (
	"time"

	"github.com/mantil-io/mantil/workspace"
)

type DataRequest struct {
	Bucket      string
	ProjectName string
	StageName   string
}

type DataResponse struct {
	Stage *workspace.Stage
}

// TODO: osiromasi ovo na samo instristicne tipove da moze ispasti referenca gore
type DeployRequest struct {
	ProjectName           string
	Stage                 *workspace.Stage
	InfrastructureChanged bool
	UpdatedFunctions      []string
	Account               *workspace.Account
	ResourceTags          map[string]string
}

type DeployResponse struct {
	Rest          string
	Ws            string
	PublicBuckets map[string]string
}

type DestroyRequest struct {
	Bucket      string
	ProjectName string
	StageName   string
}

const (
	BucketQueryParam      = "bucket"
	ProjectNameQueryParam = "projectName"
	StageNameQueryParam   = "stageName"
	CliRoleQueryParam     = "cliRole"
)

type SecurityRequest struct {
	Bucket      string
	ProjectName string
	StageName   string
	CliRole     string
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

type SetupResponse struct {
	APIGatewayRestURL string
	APIGatewayWsURL   string
	CliRole           string
}
