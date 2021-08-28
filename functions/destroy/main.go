package main

import (
	"github.com/mantil-io/mantil/api/destroy"
	"github.com/mantil-io/mantil.go"
)

func main() {
	var api = destroy.New()
	mantil.LambdaHandler(api)
}
