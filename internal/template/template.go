package template

import (
	"bytes"
	"io/ioutil"
	"strings"
	"text/template"
)

func Exec(templatePath string, data interface{}, resultPath string) error {
	resultContent, err := Render(templatePath, data)
	if err != nil {
		return err
	}
	if err := ioutil.WriteFile(resultPath, resultContent, 0644); err != nil {
		return err
	}
	return nil
}

func Render(templatePath string, data interface{}) ([]byte, error) {
	templateContent, err := ioutil.ReadFile(templatePath)
	if err != nil {
		return nil, err
	}
	funcs := template.FuncMap{"join": strings.Join}
	tpl := template.Must(template.New("").Funcs(funcs).Parse(string(templateContent)))
	buf := bytes.NewBuffer(nil)
	if err := tpl.Execute(buf, data); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
