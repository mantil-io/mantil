package controller

type function struct {
	Name       string
	ImportPath string
}

type method struct {
	Name         string
	FunctionName string
}

type test struct {
	Name       string
	ImportPath string
	Methods    []string
}

var apiDefaultTemplate = `
package {{ .Name | toLower }}

import (
	"context"
)

type {{ .Name | title }} struct{}

type DefaultRequest struct{}
type DefaultResponse struct{}

func New() *{{ .Name | title }} {
	return &{{ .Name | title }}{}
}

func ({{ .Name | first | toLower }} *{{ .Name | title }}) Default(ctx context.Context, req *DefaultRequest) (*DefaultResponse, error) {
	panic("not implemented")
}
`

var apiMethodTemplate = `
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

var apiFunctionMainTemplate = `
// Code generated by mantil DO NOT EDIT
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

var apiFunctionTestInit = `
package main

import (
        "log"
        "os"
        "os/exec"
)

var apiURL string

func init() {
        if val, ok := os.LookupEnv("MANTIL_API_URL"); ok {
                apiURL = val
                return
        }
        out, err := exec.Command("mantil", "env", "-u").Output()
        if err != nil {
                ui.Fatalf("can't find api url, execute of mantil env -u failed %v", err)
        }
        apiURL = string(out)
}
`

var apiFunctionTestTemplate = `
package main

import (
	"net/http"
	"testing"

	"github.com/gavv/httpexpect"
	"{{ .ImportPath }}/api/{{ .Name | toLower }}"
)

func Test{{ .Name | toLower | title }}(t *testing.T) {
	api := httpexpect.New(t, apiURL)

	req := {{ .Name | toLower }}.DefaultRequest {
		// TODO add attributes
	}
	api.POST("/{{ .Name | toLower }}").
			WithJSON(req).
			Expect().
			ContentType("application/json").
			Status(http.StatusOK).
			JSON().Object().
			Value("TODO")

{{ range $method := .Methods }}
	{{ $method | toLower }}Req := {{ $.Name | toLower }}.{{ $method | toLower | title }}Request {
		// TODO add attributes
	}
	api.POST("/{{ $.Name | toLower }}/{{ $method | toLower }}").
			WithJSON({{ $method | toLower }}Req).
			Expect().
			ContentType("application/json").
			Status(http.StatusOK).
			JSON().Object().
			Value("TODO")

{{ end }}
}
`