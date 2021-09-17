package cmd

import (
	"fmt"
	"os"

	"github.com/mantil-io/mantil/internal/aws"
	"github.com/mantil-io/mantil/internal/cli/commands/setup"
	"github.com/mantil-io/mantil/internal/cli/log"
	"github.com/spf13/cobra"
)

// setupCmd represents the setup command
var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Setups mantil backend infrastructure in specified AWS account",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		var awsClient *aws.AWS
		var err error
		if cmd.Flags().Changed("aws-access-key-id") {
			awsClient, err = awsFromAccessKeys(cmd)
		} else if cmd.Flags().Changed("aws-env") {
			awsClient, err = awsFromEnv(cmd)
		} else if cmd.Flags().Changed("aws-profile") {
			awsClient, err = awsFromProfile(cmd)
		}
		if err != nil {
			log.Fatal(err)
		}
		if awsClient == nil {
			log.Fatalf("could not initialize aws client")
		}
		var accountName string
		if len(args) > 0 {
			accountName = args[0]
		}
		b := setup.New(awsClient, accountName)
		destroy, err := cmd.Flags().GetBool("destroy")
		if err != nil {
			log.Fatal(err)
		}
		if err := b.Setup(destroy); err != nil {
			log.Fatal(err)
		}
	},
}

func awsFromAccessKeys(cmd *cobra.Command) (*aws.AWS, error) {
	accessKeyID, err := cmd.Flags().GetString("aws-access-key-id")
	if err != nil {
		return nil, err
	}
	if accessKeyID == "" {
		return nil, fmt.Errorf("access key id not provided")
	}
	secretAccessKey, err := cmd.Flags().GetString("aws-secret-access-key")
	if err != nil {
		return nil, err
	}
	if secretAccessKey == "" {
		return nil, fmt.Errorf("secret access key not provided")
	}
	return aws.NewWithCredentials(accessKeyID, secretAccessKey, "")
}

func awsFromEnv(cmd *cobra.Command) (*aws.AWS, error) {
	accessKeyID, ok := os.LookupEnv("AWS_ACCESS_KEY_ID")
	if !ok || accessKeyID == "" {
		return nil, fmt.Errorf("access key id not provided")
	}
	secretAccessKey, ok := os.LookupEnv("AWS_SECRET_ACCESS_KEY")
	if !ok || secretAccessKey == "" {
		return nil, fmt.Errorf("secret access key not provided")
	}
	return aws.NewWithCredentials(accessKeyID, secretAccessKey, "")
}

func awsFromProfile(cmd *cobra.Command) (*aws.AWS, error) {
	profile, err := cmd.Flags().GetString("aws-profile")
	if err != nil {
		return nil, err
	}
	if profile == "" {
		return nil, fmt.Errorf("profile not provided")
	}
	return aws.NewFromProfile(profile)
}

func init() {
	rootCmd.AddCommand(setupCmd)
	setupCmd.Flags().BoolP("destroy", "d", false, "destroy all resources created by Setup")
	setupCmd.Flags().String("aws-access-key-id", "", "access key ID for the AWS account, must be used with the aws-secret-access-key flag")
	setupCmd.Flags().String("aws-secret-access-key", "", "secret access key for the AWS account, must be used with the aws-access-key-id flag")
	setupCmd.Flags().Bool("aws-env", false, "use AWS_ACCESS_KEY_ID and AWS_SECRET_ACCESS_KEY environment variables for AWS authentication")
	setupCmd.Flags().String("aws-profile", "", "use the given profile for AWS authentication")
}
