package texts

import (
	"bytes"
	_ "embed"
	"text/template"
)

const MailFrom = "hello@mantil.com"
const ActivationMailSubject = "Mantil activation instructions"
const WelcomeMailSubject = "Welcome to Mantil!"

//go:embed activationMailBody
var activationMailBodyTemplate string

//go:embed activation.html
var activationMailHTMLBodyTemplate string

//go:embed welcome.html
var welcomeMailBodyTemplate string

func ActivationMailBody(name, activationCode string) (string, error) {
	data := struct {
		Name           string
		ActivationCode string
	}{name, activationCode}

	return renderTemplate(data, activationMailBodyTemplate)
}

const NotActivatedError = `Mantil is not activated. Please fill out the short survey at
www.mantil.com to receive your activation code.`

func WelcomeMailHTMLBody(name string) (string, error) {
	data := struct {
		Name string
	}{name}
	return renderTemplate(data, welcomeMailBodyTemplate)
}

func renderTemplate(data interface{}, content string) (string, error) {
	tpl, err := template.New("").Parse(content)
	if err != nil {
		return "", err
	}
	buf := bytes.NewBuffer(nil)
	if err := tpl.Execute(buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func ActivationHTMLMailBody(name, activationCode string) (string, error) {
	data := struct {
		Name           string
		ActivationCode string
	}{name, activationCode}

	return renderTemplate(data, activationMailHTMLBodyTemplate)
}
