package main

import (
	"github.com/mantil-io/mantil/api/security"
	"github.com/mantil-io/mantil.go"
)

func main() {
	var api = security.New()
	mantil.LambdaHandler(api)
}