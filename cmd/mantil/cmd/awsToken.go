package cmd

import (
	"bytes"
	"fmt"
	"log"
	"text/template"

	"github.com/atoz-technology/mantil-cli/internal/assets"
	"github.com/atoz-technology/mantil-cli/internal/aws"
	"github.com/spf13/cobra"
)

// awsTokenCmd represents the awsToken command
var awsTokenCmd = &cobra.Command{
	Use: "awsToken",
	Run: func(cmd *cobra.Command, args []string) {
		p, _, _ := findProject(args)
		policyTpl, err := assets.Asset("aws/project-policy.json")
		if err != nil {
			log.Fatal(err)
		}
		tpl := template.Must(template.New("").Parse(string(policyTpl)))
		buf := bytes.NewBuffer(nil)
		if err := tpl.Execute(buf, p); err != nil {
			log.Fatal(err)
		}
		awsClient, err := aws.New()
		if err != nil {
			log.Fatal(err)
		}
		c, err := awsClient.GetProjectToken(p.Name, buf.String())
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(buf.String())
		fmt.Printf("export AWS_ACCESS_KEY_ID=%s\n", *c.AccessKeyId)
		fmt.Printf("export AWS_SECRET_ACCESS_KEY=%s\n", *c.SecretAccessKey)
		fmt.Printf("export AWS_SESSION_TOKEN=%s\n", *c.SessionToken)
	},
}

func init() {
	rootCmd.AddCommand(awsTokenCmd)
}
