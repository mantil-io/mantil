package destroy

import (
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
		return err
	}
	os.RemoveAll(name)
	aws, err := aws.New()
	if err != nil {
		log.Fatal(err)
	}
	bucketName := project.Bucket
	bucketExists, _ := aws.S3BucketExists(bucketName)
	if bucketExists {
		err = aws.DeleteS3Bucket(bucketName)
		if err != nil {
			log.Fatal(err)
		}
	}
	log.Printf("Deleting github repository...")
	ghClient, err := github.NewClient()
	if err != nil {
		log.Fatal(err)
	}
	err = ghClient.DeleteRepo(name)
	if err != nil {
		log.Fatal(err)
	}
	return nil
}
