package cmd

import (
	"bytes"
	"fmt"
	"io/fs"
	"io/ioutil"
	"path"
	"strings"
	"text/template"

	"github.com/mantil-io/mantil/cli/ui"
	"github.com/spf13/cobra"
)

// markdown documentation generator
type mdGenerator struct {
	dir          string            // template and ouput dir
	descriptions map[string]string // filename to description map
	dryRun       bool
}

type mdData struct {
	Use  string
	Help string
}

func (g *mdGenerator) gen(rootCmd *cobra.Command) error {
	g.descriptions = make(map[string]string)
	if err := g.genForCmdAndSub(rootCmd); err != nil {
		return err
	}
	fmt.Printf("\nDescriptions for readme file:\n")
	for fn, desc := range g.descriptions {
		fmt.Printf("%-24s\t%s\n", fn, desc)
	}
	return nil
}

func (g *mdGenerator) genForCmdAndSub(cmd *cobra.Command) error {
	if err := g.genForCmd(cmd); err != nil {
		return err
	}
	for _, subCmd := range cmd.Commands() {
		if err := g.genForCmdAndSub(subCmd); err != nil {
			return err
		}
	}
	return nil
}

func (g *mdGenerator) genForCmd(cmd *cobra.Command) error {
	basename := strings.Replace(cmd.CommandPath(), " ", "_", -1)
	mdfile := basename + ".md"
	outputFile := path.Join(g.dir, mdfile)
	data := mdData{cmd.CommandPath(), g.help(cmd)}
	mdContent, err := g.runTemplate(data)
	if err != nil {
		return err
	}
	g.descriptions[mdfile] = cmd.Short
	if !g.dryRun {
		if err := ioutil.WriteFile(outputFile, mdContent, fs.ModePerm); err != nil {
			return err
		}
	}
	ui.Info("created %s", outputFile)
	return nil
}

func (d mdGenerator) help(cmd *cobra.Command) string {
	bb := bytes.NewBuffer(nil)
	cmd.SetOutput(bb)
	_ = cmd.Help()

	buf := bb.Bytes()
	buf = bytes.Replace(buf, []byte("<"), []byte("&lt;"), -1)
	buf = bytes.Replace(buf, []byte(">"), []byte("&gt;"), -1)
	// replace bold headers with ###
	// enclose other content into <pre></pre>
	buf = bytes.Replace(buf, []byte(bold), []byte("### "), 1)
	buf = bytes.Replace(buf, []byte(bold), []byte("</pre>\n### "), -1)
	buf = bytes.Replace(buf, []byte(clear), []byte("\n<pre>"), -1)
	buf = append(buf, []byte("</pre>")...)
	buf = bytes.Replace(buf, []byte("\n</pre>"), []byte("</pre>"), -1)

	return string(buf)
}

var commandTemplate = `
# {{.Use}}

{{.Help}}
`

func (d mdGenerator) runTemplate(data mdData) ([]byte, error) {
	// note use text/template package html/template
	tpl, err := template.New("").Parse(commandTemplate)
	if err != nil {
		return nil, err
	}
	out := bytes.NewBuffer(nil)
	if err := tpl.Execute(out, data); err != nil {
		return nil, err
	}
	return out.Bytes(), nil
}
