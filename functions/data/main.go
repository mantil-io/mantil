package main

import (
	"github.com/mantil-io/mantil/api/data"
	"github.com/mantil-io/mantil.go"
)

func main() {
	var api = data.New()
	mantil.LambdaHandler(api)
}
