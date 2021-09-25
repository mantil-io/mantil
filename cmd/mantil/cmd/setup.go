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
	Use:   "setup [account-name]",
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
		} else {
			log.Fatalf("aws profile not provided, check `mantil setup --help` for more info")
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

		b := setup.New(awsClient, version, accountName)
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
	region, err := cmd.Flags().GetString("aws-region")
	if err != nil {
		return nil, err
	}
	if region == "" {
		return nil, fmt.Errorf("region not provided")
	}
	return aws.NewWithCredentials(accessKeyID, secretAccessKey, "", region)
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
	region, ok := os.LookupEnv("AWS_DEFAULT_REGION")
	if !ok || region == "" {
		return nil, fmt.Errorf("region not provided")
	}
	return aws.NewWithCredentials(accessKeyID, secretAccessKey, "", region)
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
	setupCmd.Flags().String("aws-access-key-id", "", "access key ID for the AWS account, must be used with the aws-secret-access-key and aws-region flags")
	setupCmd.Flags().String("aws-secret-access-key", "", "secret access key for the AWS account, must be used with the aws-access-key-id and aws-region flags")
	setupCmd.Flags().String("aws-region", "", "region for the AWS account, must be used with and aws-access-key-id and aws-secret-access-key flags")
	setupCmd.Flags().Bool("aws-env", false, "use AWS_ACCESS_KEY_ID, AWS_SECRET_ACCESS_KEY and AWS_DEFAULT_REGION environment variables for AWS authentication")
	setupCmd.Flags().String("aws-profile", "", "use the given profile for AWS authentication")
}
