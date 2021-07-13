package cmd

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path"
	"strings"
	"text/template"

	"github.com/atoz-technology/mantil-cli/internal/assets"
	"github.com/atoz-technology/mantil-cli/internal/aws"
	"github.com/atoz-technology/mantil-cli/internal/terraform"
	"github.com/atoz-technology/mantil-cli/pkg/mantil"
	"github.com/atoz-technology/mantil-cli/pkg/shell"
	"github.com/spf13/cobra"
)

func createProject(name, funcsPath string) mantil.Project {
	project := mantil.NewProject(name)
	files, err := ioutil.ReadDir(funcsPath)
	if err != nil {
		log.Fatal(err)
	}
	// go through functions in functions directory
	for _, f := range files {
		if !f.IsDir() {
			continue
		}
		name := f.Name()
		f := mantil.Function{
			Path:       name,
			Name:       name,
			S3Key:      fmt.Sprintf("functions/%s.zip", name),
			Runtime:    "go1.x",
			MemorySize: 128,
			Timeout:    60 * 15,
			Handler:    name,
		}
		f.URL = fmt.Sprintf("https://%s/%s/%s", project.Organization.DNSZone, project.Name, f.Path)
		project.Functions = append(project.Functions, f)
	}
	return project
}

func renderTerraformTemplate(path string, project mantil.Project) error {
	funcs := template.FuncMap{"join": strings.Join}
	tfTpl, err := assets.Asset("terraform/templates/main.tf")
	if err != nil {
		return err
	}
	tpl := template.Must(template.New("").Funcs(funcs).Parse(string(tfTpl)))
	buf := bytes.NewBuffer(nil)
	if err := tpl.Execute(buf, project); err != nil {
		return err
	}
	if err := ioutil.WriteFile(path, buf.Bytes(), 0644); err != nil {
		return err
	}
	return nil
}

// deployCmd represents the deploy command
var deployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "Creates infrastructure and deploys updates to lambda functions",
	Run: func(cmd *cobra.Command, args []string) {
		awsClient, err := aws.New()
		if err != nil {
			log.Fatalf("error while initialising aws - %v", err)
		}
		projectRoot, err := os.Getwd()
		if err != nil {
			log.Fatal(err)
		}
		projectName := path.Base(projectRoot)
		project := createProject(projectName, "functions")
		for _, f := range project.Functions {
			name := f.Name
			log.Printf("deploying function %s", name)
			funcDir := fmt.Sprintf("functions/%s", name)
			if err := shell.Exec([]string{"env", "GOOS=linux", "GOARCH=amd64", "go", "build", "-o", name}, funcDir); err != nil {
				log.Fatalf("skipping function %s due to error while building binary - %v", name, err)
			}
			buf, err := createZipFunction(fmt.Sprintf("%s/%s", funcDir, name), name)
			if err != nil {
				log.Fatal(err)
			}
			if err := awsClient.PutObjectToS3Bucket(project.Bucket, f.S3Key, bytes.NewReader(buf)); err != nil {
				log.Fatalf("error while uploading function %s to S3 - %v", f.Name, err)
			}
		}
		renderTerraformTemplate("main.tf", project)
		tf := terraform.New(".")
		if err := tf.Init(); err != nil {
			log.Fatal(err)
		}
		if err := tf.Plan(false); err != nil {
			log.Fatal(err)
		}
		if err := tf.Apply(false); err != nil {
			log.Fatal(err)
		}
		for _, f := range project.Functions {
			lambdaName := fmt.Sprintf("%s-mantil-team-%s-%s", project.Organization.Name, project.Name, f.Name)
			if err := awsClient.UpdateLambdaFunctionCodeFromS3(lambdaName, project.Bucket, f.S3Key); err != nil {
				log.Fatal(err)
			}
		}
	},
}

func createZipFunction(path, name string) ([]byte, error) {
	buf := new(bytes.Buffer)
	w := zip.NewWriter(buf)

	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return nil, err
	}

	hdr, err := zip.FileInfoHeader(info)
	if err != nil {
		return nil, err
	}

	// using base name in the header so zip doesn't create a directory
	hdr.Name = name
	hdr.Method = zip.Deflate
	dst, err := w.CreateHeader(hdr)
	if err != nil {
		return nil, err
	}

	_, err = io.Copy(dst, file)
	if err != nil {
		return nil, err
	}

	if err := w.Close(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func init() {
	rootCmd.AddCommand(deployCmd)
}
