package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/atoz-technology/mantil-cli/internal/aws"
	"github.com/atoz-technology/mantil-cli/internal/github"
	"github.com/atoz-technology/mantil-cli/internal/terraform"
	"github.com/atoz-technology/mantil-cli/pkg/mantil"
	"github.com/spf13/cobra"
)

// destroyCmd represents the destroy command
var destroyCmd = &cobra.Command{
	Use:  "destroy",
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]
		_, err := os.Stat(name)
		if err == nil {
			fmt.Println("Destroying infrastructure...")
			tf := terraform.New(name)
			if err := tf.Init(); err != nil {
				log.Fatal(err)
			}
			if err := tf.Plan(true); err != nil {
				log.Fatal(err)
			}
			if err := tf.Apply(true); err != nil {
				log.Fatal(err)
			}
			os.RemoveAll(name)
		}
		aws, err := aws.New()
		if err != nil {
			log.Fatal(err)
		}
		bucketName := mantil.NewProject(name).Bucket
		bucketExists, _ := aws.S3BucketExists(bucketName)
		if bucketExists {
			err = aws.DeleteS3Bucket(bucketName)
			if err != nil {
				log.Fatal(err)
			}
		}
		fmt.Println("Deleting github repository...")
		ghClient, err := github.NewClient()
		if err != nil {
			log.Fatal(err)
		}
		err = ghClient.DeleteRepo(name)
		if err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(destroyCmd)
}
