package main

import (
	"github.com/atoz-technology/mantil-backend/api/data"
	"github.com/atoz-technology/mantil.go"
)

func main() {
	var api = data.New()
	mantil.LambdaHandler(api)
}
