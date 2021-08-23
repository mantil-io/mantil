package main

import (
	"github.com/mantil-io/mantil-backend/api/initialize"
	"github.com/mantil-io/mantil.go"
)

func main() {
	var api = initialize.New()
	mantil.LambdaHandler(api)
}
