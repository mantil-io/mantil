package deploy

import (
	"bytes"
	"fmt"
	"log"
	"path"

	"github.com/atoz-technology/mantil-cli/internal/aws"
	"github.com/atoz-technology/mantil-cli/internal/mantil"
	"github.com/atoz-technology/mantil-cli/internal/shell"
	"github.com/atoz-technology/mantil-cli/internal/terraform"
	"github.com/atoz-technology/mantil-cli/internal/util"
)

type DeployCmd struct {
	aws     *aws.AWS
	project *mantil.Project
}

func New(projectRoot string) (*DeployCmd, error) {
	awsClient, err := aws.New()
	if err != nil {
		return nil, err
	}

	projectName := path.Base(projectRoot)
	project, err := mantil.NewProject(projectName, "functions")
	return &DeployCmd{
		aws:     awsClient,
		project: project,
	}, nil
}

func (d *DeployCmd) UploadFunctions() error {
	for _, f := range d.project.Functions {
		name := f.Name
		log.Printf("uploading function %s", name)
		funcDir := fmt.Sprintf("functions/%s", name)
		if err := shell.Exec([]string{"env", "GOOS=linux", "GOARCH=amd64", "go", "build", "-o", name}, funcDir); err != nil {
			log.Printf("skipping function %s due to error while building binary - %v", name, err)
			continue
		}
		buf, err := util.CreateZipForFile(fmt.Sprintf("%s/%s", funcDir, name), name)
		if err != nil {
			log.Printf("skipping function %s due to error while zipping binary - %v", name, err)
			continue
		}
		if err := d.aws.PutObjectToS3Bucket(d.project.Bucket, f.S3Key, bytes.NewReader(buf)); err != nil {
			log.Printf("skipping function %s due to error while uploading to s3 - %v", name, err)
			continue
		}
	}
	return nil
}

func (d *DeployCmd) CreateInfrastructure() error {
	tf := terraform.New(".")
	if err := tf.ApplyForProject(d.project, false); err != nil {
		return fmt.Errorf("error while applying terraform for project %s - %v", d.project.Name, err)
	}
	return nil
}

func (d *DeployCmd) UpdateLambdaFunctions() error {
	for _, f := range d.project.Functions {
		lambdaName := fmt.Sprintf("%s-mantil-team-%s-%s", d.project.Organization.Name, d.project.Name, f.Name)
		if err := d.aws.UpdateLambdaFunctionCodeFromS3(lambdaName, d.project.Bucket, f.S3Key); err != nil {
			log.Printf("skipping updating lambda %s due to error %v", lambdaName, err)
		}
	}
	return nil
}
