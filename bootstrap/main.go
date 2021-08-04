package main

import (
	"github.com/atoz-technology/mantil-backend/api/bootstrap"
	"github.com/atoz-technology/mantil.go"
)

func main() {
	var api = bootstrap.New()
	mantil.LambdaHandler(api)
}
