package main

import (
	"github.com/mantil-io/mantil.go"
	"github.com/mantil-io/mantil/node/api/setup"
)

func main() {
	var api = setup.New()
	mantil.LambdaHandler(api)
}
