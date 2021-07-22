package main

import (
	"github.com/atoz-technology/mantil-backend/api/hello"
	"github.com/atoz-technology/mantil.go"
)

func main() {
	var api = hello.New()
	mantil.LambdaHandler(api)
}
