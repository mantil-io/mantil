package cmd

import (
	"fmt"
	"os"

	"github.com/mantil-io/mantil/aws"
	"github.com/mantil-io/mantil/cli/commands/setup"
	"github.com/mantil-io/mantil/cli/log"
	"github.com/spf13/cobra"
)

// setupCmd represents the setup command
var setupCmd = &cobra.Command{
	Use: "setup [account-name]",
	// TODO: objasni u da se ocekuje jedna od tri variajante aws credentials
	//       objasni da access i secret idu u paru
	Short: "Setups mantil backend infrastructure in specified AWS account",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		destroy, err := cmd.Flags().GetBool("destroy")
		if err != nil {
			log.Fatal(err)
		}
		var accountName string
		if len(args) > 0 {
			accountName = args[0]
		}

		awsClient, err := createAwsClient(cmd)
		if err != nil {
			return err
		}

		v, ok := setup.GetVersion(cmd.Context())
		if !ok {
			return fmt.Errorf("version not found in context")
		}
		b := setup.New(awsClient, v, accountName)
		if destroy {
			err = b.Destroy()
		} else {
			err = b.Create()
		}
		return err
	},
}

func createAwsClient(cmd *cobra.Command) (*aws.AWS, error) {
	var c credentials
	if err := c.read(cmd); err != nil {
		return nil, err
	}
	if c.profile != "" {
		return aws.NewFromProfile(c.profile)
	}
	return aws.NewWithCredentials(c.accessKeyID, c.secretAccessKey, "", c.region)
}

type credentials struct {
	accessKeyID     string
	secretAccessKey string
	region          string
	profile         string
}

func (c *credentials) read(cmd *cobra.Command) error {
	var err error
	if cmd.Flags().Changed("aws-access-key-id") {
		err = c.awsFromAccessKeys(cmd)
	} else if cmd.Flags().Changed("aws-env") {
		err = c.awsFromEnv(cmd)
	} else if cmd.Flags().Changed("aws-profile") {
		err = c.awsFromProfile(cmd)
	} else {
		err = fmt.Errorf("aws credentials not provided")
	}
	if err != nil {
		return err
	}
	return nil
}

func (c *credentials) awsFromAccessKeys(cmd *cobra.Command) error {
	var err error
	c.accessKeyID, err = cmd.Flags().GetString("aws-access-key-id")
	if err != nil {
		return err
	}
	if c.accessKeyID == "" {
		return fmt.Errorf("access key id not provided")
	}
	c.secretAccessKey, err = cmd.Flags().GetString("aws-secret-access-key")
	if err != nil {
		return err
	}
	if c.secretAccessKey == "" {
		return fmt.Errorf("secret access key not provided")
	}
	c.region, err = cmd.Flags().GetString("aws-region")
	if err != nil {
		return err
	}
	if c.region == "" {
		return fmt.Errorf("region not provided")
	}
	return nil
}

func (c *credentials) awsFromEnv(cmd *cobra.Command) error {
	var ok bool
	c.accessKeyID, ok = os.LookupEnv("AWS_ACCESS_KEY_ID")
	if !ok || c.accessKeyID == "" {
		return fmt.Errorf("access key id not provided")
	}
	c.secretAccessKey, ok = os.LookupEnv("AWS_SECRET_ACCESS_KEY")
	if !ok || c.secretAccessKey == "" {
		return fmt.Errorf("secret access key not provided")
	}
	c.region, ok = os.LookupEnv("AWS_DEFAULT_REGION")
	if !ok || c.region == "" {
		return fmt.Errorf("region not provided")
	}
	return nil
}

func (c *credentials) awsFromProfile(cmd *cobra.Command) error {
	var err error
	c.profile, err = cmd.Flags().GetString("aws-profile")
	if err != nil {
		return err
	}
	if c.profile == "" {
		return fmt.Errorf("profile not provided")
	}
	return nil
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
