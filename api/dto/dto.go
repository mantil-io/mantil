package dto

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
