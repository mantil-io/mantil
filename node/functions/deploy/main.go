package main

import (
	"github.com/mantil-io/mantil.go"
	"github.com/mantil-io/mantil/node/api/deploy"
)

func main() {
	var api = deploy.New()
	mantil.LambdaHandler(api)
}
