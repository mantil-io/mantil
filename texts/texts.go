package texts

import (
	"bytes"
	_ "embed"
	"text/template"
)

const ActivationMailFrom = "hello@mantil.com"
const ActivationMailSubject = "Mantil activation instructions"

//go:embed activationMailBody
var activationMailBodyTemplate string

func ActivationMailBody(name, activationCode string) (string, error) {
	data := struct {
		Name           string
		ActivationCode string
	}{name, activationCode}

	tpl, err := template.New("").Parse(activationMailBodyTemplate)
	if err != nil {
		return "", err
	}
	buf := bytes.NewBuffer(nil)
	if err := tpl.Execute(buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}
