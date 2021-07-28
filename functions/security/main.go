package main

import (
	"github.com/atoz-technology/mantil-backend/api/security"
	"github.com/atoz-technology/mantil.go"
)

func main() {
	var api = security.New()
	mantil.LambdaHandler(api)
}
