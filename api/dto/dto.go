package dto

import (
	"time"

	"github.com/mantil-io/mantil/workspace"
)

type DataRequest struct {
	ProjectName string
	StageName   string
}

type DataResponse struct {
	Stage *workspace.Stage
}

type DeployRequest struct {
	ProjectName string
	Stage       *workspace.Stage
	Account     *workspace.Account
}

type DestroyRequest struct {
	ProjectName string
	StageName   string
}

const (
	ProjectNameQueryParam = "projectName"
	StageNameQueryParam   = "stageName"
	CliRoleQueryParam     = "cliRole"
)

type SecurityRequest struct {
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
	FunctionsBucket string
	FunctionsPath   string
	PublicKey       string
	ResourceSuffix  string
	Destroy         bool
}

type SetupResponse struct {
	APIGatewayRestURL string
	APIGatewayWsURL   string
	CliRole           string
}
