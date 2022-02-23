package main

import (
	"github.com/mantil-io/mantil.go"
	"github.com/mantil-io/mantil/node/api/auth"
)

func main() {
	var api = auth.New()
	mantil.LambdaHandler(api)
}
