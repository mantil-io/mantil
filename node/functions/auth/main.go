package main

import (
	"github.com/mantil-io/mantil.go"
	"github.com/mantil-io/mantil/node/api/node"
)

func main() {
	var api = node.NewAuth()
	mantil.LambdaHandler(api)
}
