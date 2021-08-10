package generate

import (
	"bytes"
	"html/template"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func GenerateFromTemplate(tplDef string, data interface{}, outPath string) error {
	out, err := runTemplate(tplDef, data)
	if err != nil {
		return err
	}
	out, err = format(out)
	if err != nil {
		return err
	}
	return save(out, outPath)
}

func runTemplate(tplDef string, data interface{}) ([]byte, error) {
	fcs := template.FuncMap{
		"join":    strings.Join,
		"toLower": strings.ToLower,
		"title":   strings.Title,
	}
	tpl := template.Must(template.New("").Funcs(fcs).Parse(tplDef))
	buf := bytes.NewBuffer(nil)
	if err := tpl.Execute(buf, data); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func format(in []byte) ([]byte, error) {
	cmd := exec.Command("gofmt")
	cmd.Stdin = strings.NewReader(string(in))
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	return out, nil
}

func save(in []byte, path string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return err
	}
	if err := ioutil.WriteFile(path, in, 0644); err != nil {
		return err
	}
	return nil
}
