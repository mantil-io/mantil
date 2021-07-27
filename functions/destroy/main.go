package main

import (
	"github.com/atoz-technology/mantil-backend/api/destroy"
	"github.com/atoz-technology/mantil.go"
)

func main() {
	var api = destroy.New()
	mantil.LambdaHandler(api)
}
