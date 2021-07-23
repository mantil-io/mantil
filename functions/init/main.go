package main

import (
	"github.com/atoz-technology/mantil-backend/api/initialize"
	"github.com/atoz-technology/mantil.go"
)

func main() {
	var api = initialize.New()
	mantil.LambdaHandler(api)
}
