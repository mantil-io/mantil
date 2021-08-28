package main

import (
	"github.com/mantil-io/mantil/api/initialize"
	"github.com/mantil-io/mantil.go"
)

func main() {
	var api = initialize.New()
	mantil.LambdaHandler(api)
}
