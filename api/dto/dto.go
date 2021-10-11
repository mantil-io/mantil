package dto

import "github.com/mantil-io/mantil/workspace"

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

type SecurityRequest struct {
	ProjectName string
	StageName   string
	CliRole     string
}

type SecurityResponse struct {
	AccessKeyID     string
	SecretAccessKey string
	SessionToken    string
	Region          string
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
