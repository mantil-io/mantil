package main

import (
	"github.com/mantil-io/mantil-backend/api/security"
	"github.com/mantil-io/mantil.go"
)

func main() {
	var api = security.New()
	mantil.LambdaHandler(api)
}
