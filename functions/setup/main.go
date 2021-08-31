package main

import (
	"github.com/mantil-io/mantil/api/setup"
	"github.com/mantil-io/mantil.go"
)

func main() {
	var api = setup.New()
	mantil.LambdaHandler(api)
}