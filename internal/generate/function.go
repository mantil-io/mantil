package generate

type Function struct {
	Name       string
	ImportPath string
}

type Method struct {
	Name         string
	FunctionName string
}

var APIDefaultTemplate = `
package {{ .Name | toLower }}

import (
	"context"
)

type {{ .Name | title }} struct{}

type DefaultRequest struct{}
type DefaultResponse struct{}

func ({{ .Name | first | toLower }} *{{ .Name | title }}) Init(ctx context.Context) {}

func ({{ .Name | first | toLower }} *{{ .Name | title }}) Default(ctx context.Context, req *DefaultRequest) (*DefaultResponse, error) {
	panic("not implemented")
}

func New() *{{ .Name | title }} {
	return &{{ .Name | title }}{}
}
`

var APIMethodTemplate = `
package {{ .FunctionName | toLower }}

import (
	"context"
)

type {{ .Name | title }}Request struct{}
type {{ .Name | title }}Response struct{}

func ({{ .FunctionName | first | toLower }} *{{ .FunctionName | title }}) {{ .Name | title }}(ctx context.Context, req *{{ .Name | title }}Request) (*{{ .Name | title }}Response, error) {
	panic("not implemented")
}
`

var APIFunctionMainTemplate = `
package main

import (
	"{{ .ImportPath }}/api/{{ .Name | toLower }}"
	"github.com/mantil-io/mantil.go"
)

func main() {
	var api = {{ .Name | toLower }}.New()
	mantil.LambdaHandler(api)
}
`
