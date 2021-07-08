package cmd

import (
	"log"
	"os"

	"github.com/atoz-technology/mantil-cli/internal/aws"
	"github.com/atoz-technology/mantil-cli/internal/github"
	"github.com/atoz-technology/mantil-cli/pkg/mantil"
	"github.com/spf13/cobra"
)

// destroyCmd represents the destroy command
var destroyCmd = &cobra.Command{
	Use:  "destroy",
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]
		os.RemoveAll(name)
		awsSdk, err := aws.NewSDK()
		if err != nil {
			log.Fatal(err)
		}
		bucketName := mantil.NewProject(name).Bucket
		bucketExists, _ := awsSdk.S3BucketExists(bucketName)
		if bucketExists {
			err = awsSdk.DeleteS3Bucket(bucketName)
			if err != nil {
				log.Fatal(err)
			}
		}
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
