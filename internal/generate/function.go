package generate

type Function struct {
	Name       string
	ImportPath string
}

type Method struct {
	Name         string
	FunctionName string
}

var APIRootTemplate = `
package {{ .Name | toLower }}

import (
	"context"
)

type {{ .Name | title }} struct{}

type RootRequest struct{}
type RootResponse struct{}

func ({{ .Name | first | toLower }} *{{ .Name | title }}) Init(ctx context.Context) {}

func ({{ .Name | first | toLower }} *{{ .Name | title }}) Invoke(ctx context.Context, req *RootRequest) (*RootResponse, error) {
	return {{ .Name | first | toLower }}.Root(ctx, req)
}

func ({{ .Name | first | toLower }} *{{ .Name | title }}) Root(ctx context.Context, req *RootRequest) (*RootResponse, error) {
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
