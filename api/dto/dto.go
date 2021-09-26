package dto

import "github.com/mantil-io/mantil/config"

type DataRequest struct {
	ProjectName string
}

type DataResponse struct {
	Project *config.Project
}

type SecurityRequest struct {
	ProjectName string
	StageName   string
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
	Destroy         bool
}

type SetupResponse struct {
	APIGatewayRestURL string
	APIGatewayWsURL   string
}
