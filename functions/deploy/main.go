package main

import (
	"github.com/atoz-technology/mantil-backend/api/deploy"
	"github.com/atoz-technology/mantil.go"
)

func main() {
	var api = deploy.New()
	mantil.LambdaHandler(api)
}
