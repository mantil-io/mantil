package destroy

import (
	"fmt"
	"log"
	"os"
	"path"

	"github.com/atoz-technology/mantil-cli/internal/aws"
	"github.com/atoz-technology/mantil-cli/internal/github"
	"github.com/atoz-technology/mantil-cli/internal/mantil"
	"github.com/atoz-technology/mantil-cli/internal/terraform"
)

type DestroyCmd struct {
}

func New() *DestroyCmd {
	return &DestroyCmd{}
}

func (d *DestroyCmd) DestroyInfrastructure(name string) error {
	_, err := os.Stat(name)
	if err != nil {
		return err
	}
	log.Printf("Destroying infrastructure...")
	funcsPath := path.Join(name, "functions")
	project, err := mantil.NewProject(name, funcsPath)
	if err != nil {
		return err
	}
	tf := terraform.New(name)
	if err := tf.ApplyForProject(project, true); err != nil {
		return fmt.Errorf("could not terraform destroy - %v", err)
	}
	os.RemoveAll(name)
	aws, err := aws.New()
	if err != nil {
		return fmt.Errorf("could not initialize aws - %v", err)
	}
	bucketName := project.Bucket
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
