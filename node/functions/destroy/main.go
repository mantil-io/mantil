package main

import (
	"github.com/mantil-io/mantil.go"
	"github.com/mantil-io/mantil/node/api/destroy"
)

func main() {
	var api = destroy.New()
	mantil.LambdaHandler(api)
}
