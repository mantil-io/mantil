package main

import (
	"github.com/mantil-io/mantil-backend/api/destroy"
	"github.com/mantil-io/mantil.go"
)

func main() {
	var api = destroy.New()
	mantil.LambdaHandler(api)
}
