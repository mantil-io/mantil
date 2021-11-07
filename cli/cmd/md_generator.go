package cmd

import (
	"bytes"
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/mantil-io/mantil/cli/ui"
	"github.com/spf13/cobra"
)

// markdown documentation generator
type mdGenerator struct {
	dir string // template and ouput dir
}

type mdData struct {
	Description string
	Help        string
}

func (g mdGenerator) gen(rootCmd *cobra.Command) error {
	if err := g.genForCmd(rootCmd); err != nil {
		return err
	}
	for _, subCmd := range rootCmd.Commands() {
		if err := g.genForCmd(subCmd); err != nil {
			return err
		}
		if subCmd.HasSubCommands() {
			if err := g.gen(subCmd); err != nil {
				return err
			}
		}
	}
	return nil
}

func (g mdGenerator) genForCmd(cmd *cobra.Command) error {
	basename := strings.Replace(cmd.CommandPath(), " ", "_", -1) + ".md"
	templateFile := filepath.Join(g.dir, basename+".tmpl")
	outputFile := filepath.Join(g.dir, basename)

	content, err := ioutil.ReadFile(templateFile)
	if err != nil {
		if os.IsNotExist(err) {
			ui.Errorf("template %s not found", templateFile)
			return nil
		}
		return err
	}

	data := mdData{cmd.Short, g.help(cmd)}
	mdContent, err := g.runTemplate(content, data)
	if err != nil {
		return err
	}

	if err := ioutil.WriteFile(outputFile, mdContent, fs.ModePerm); err != nil {
		return err
	}
	ui.Info("created %s", outputFile)
	return nil
}

func (d mdGenerator) help(cmd *cobra.Command) string {
	bb := bytes.NewBuffer(nil)
	cmd.SetOutput(bb)
	_ = cmd.Help()

	buf := bb.Bytes()
	// replace bold headers with ###
	// enclose other content into <pre></pre>
	buf = bytes.Replace(buf, []byte(bold), []byte("### "), 1)
	buf = bytes.Replace(buf, []byte(bold), []byte("</pre>\n### "), -1)
	buf = bytes.Replace(buf, []byte(clear), []byte("\n<pre>"), -1)
	buf = append(buf, []byte("</pre>")...)
	buf = bytes.Replace(buf, []byte("\n</pre>"), []byte("</pre>"), -1)

	return string(buf)
}

func (d mdGenerator) runTemplate(content []byte, data mdData) ([]byte, error) {
	// note use text/template package html/template
	tpl, err := template.New("").Parse(string(content))
	if err != nil {
		return nil, err
	}
	out := bytes.NewBuffer(nil)
	if err := tpl.Execute(out, data); err != nil {
		return nil, err
	}
	return out.Bytes(), nil
}
