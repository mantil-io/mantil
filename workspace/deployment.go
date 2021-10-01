package workspace

import (
	"fmt"
	"strings"

	"github.com/mantil-io/mantil/aws"
	"gopkg.in/yaml.v2"
)

const (
	// env variable containing comma separated list of env variables
	// that should be added as tags to all resource created by deployment lambda functions
	EnvMantilDeploymentTags = "MANTIL_ENV_TAGS"
)

var (
	// list of env variables for EnvMantilDeploymentTags
	// variables specified here should be passed to all lambda functions
	MantilDeploymentTags = []string{EnvProjectName, EnvStageName}
)

func SaveDeploymentState(projectName string, stage *Stage) error {
	aws, err := aws.New()
	if err != nil {
		return err
	}
	bucket, err := Bucket(aws)
	if err != nil {
		return err
	}
	buf, err := yaml.Marshal(stage)
	if err != nil {
		return err
	}
	s3Key := DeploymentStateS3Key(projectName, stage.Name)
	if err := aws.PutObjectToS3Bucket(bucket, s3Key, buf); err != nil {
		return err
	}
	return nil
}

func LoadDeploymentState(projectName, stageName string) (*Stage, error) {
	aws, err := aws.New()
	if err != nil {
		return nil, err
	}
	bucket, err := Bucket(aws)
	if err != nil {
		return nil, err
	}
	s3Key := DeploymentStateS3Key(projectName, stageName)
	buf, err := aws.GetObjectFromS3Bucket(bucket, s3Key)
	if err != nil {
		return nil, err
	}
	s := &Stage{}
	if err := yaml.Unmarshal(buf, s); err != nil {
		return nil, err
	}
	return s, nil
}

func DeleteDeploymentState(projectName, stageName string) error {
	aws, err := aws.New()
	if err != nil {
		return err
	}
	bucket, err := Bucket(aws)
	if err != nil {
		return err
	}
	return aws.DeleteInS3Bucket(bucket, DeploymentBucketPrefix(projectName, stageName))
}

func DeploymentEnv(projectName, stageName string) map[string]string {
	env := map[string]string{
		EnvProjectName:          projectName,
		EnvStageName:            stageName,
		EnvMantilDeploymentTags: strings.Join(MantilDeploymentTags, ","),
	}
	return env
}

func CleanupResourcesFromDeployment(projectName, stageName string) error {
	awsClient, err := aws.New()
	if err != nil {
		return err
	}
	tags := []aws.TagFilter{
		{Key: EnvProjectName, Values: []string{projectName}},
		{Key: EnvStageName, Values: []string{stageName}},
	}
	if err := awsClient.DeleteDynamodbTablesByTags(tags); err != nil {
		return err
	}
	return nil
}

func DeploymentBucketPrefix(projectName, stageName string) string {
	return fmt.Sprintf("deployments/%s/%s", projectName, stageName)
}

func DeploymentStateS3Key(projectName, stageName string) string {
	return fmt.Sprintf("%s/state.yml", DeploymentBucketPrefix(projectName, stageName))
}
