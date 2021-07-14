package destroy

import (
	"fmt"
	"log"
	"os"

	"github.com/atoz-technology/mantil-cli/internal/aws"
	"github.com/atoz-technology/mantil-cli/internal/github"
	"github.com/atoz-technology/mantil-cli/internal/mantil"
	"github.com/atoz-technology/mantil-cli/internal/terraform"
)

type DestroyCmd struct {
	aws     *aws.AWS
	project *mantil.Project
}

func New(projectName string) (*DestroyCmd, error) {
	awsClient, err := aws.New()
	if err != nil {
		return nil, err
	}
	project, err := mantil.LoadProject(projectName)
	if err != nil {
		return nil, err
	}
	return &DestroyCmd{
		aws:     awsClient,
		project: project,
	}, nil
}

func (d *DestroyCmd) Destroy() error {
	name := d.project.Name
	tf := terraform.New(name)
	if err := tf.ApplyForProject(d.project, true); err != nil {
		return fmt.Errorf("could not terraform destroy - %v", err)
	}
	os.RemoveAll(name)
	aws, err := aws.New()
	if err != nil {
		return fmt.Errorf("could not initialize aws - %v", err)
	}
	bucketName := d.project.Bucket
	bucketExists, _ := aws.S3BucketExists(bucketName)
	if bucketExists {
		err = aws.DeleteS3Bucket(bucketName)
		if err != nil {
			return fmt.Errorf("could not delete bucket %s - %v", bucketName, err)
		}
	}
	log.Printf("Deleting github repository...")
	ghClient, err := github.NewClient()
	if err != nil {
		return fmt.Errorf("could not initialize github client - %v", err)
	}
	err = ghClient.DeleteRepo(name)
	if err != nil {
		return fmt.Errorf("could not delete repo %s - %v", name, err)
	}
	return nil
}
