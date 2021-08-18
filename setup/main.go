package main

import (
	"github.com/atoz-technology/mantil-backend/api/setup"
	"github.com/atoz-technology/mantil.go"
)

func main() {
	var api = setup.New()
	mantil.LambdaHandler(api)
}
