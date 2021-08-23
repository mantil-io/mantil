package main

import (
	"github.com/mantil-io/mantil-backend/api/deploy"
	"github.com/mantil-io/mantil.go"
)

func main() {
	var api = deploy.New()
	mantil.LambdaHandler(api)
}
