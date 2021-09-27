package config

import (
	"fmt"

	"github.com/mantil-io/mantil/aws"
	"gopkg.in/yaml.v2"
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

func DeploymentBucketPrefix(projectName, stageName string) string {
	return fmt.Sprintf("deployments/%s/%s", projectName, stageName)
}

func DeploymentStateS3Key(projectName, stageName string) string {
	return fmt.Sprintf("%s/state.yml", DeploymentBucketPrefix(projectName, stageName))
}
