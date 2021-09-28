package config

import (
	"encoding/json"
	"fmt"

	"github.com/mantil-io/mantil/aws"
)

const (
	s3RuntimeConfigKey = "config.json"
)

func Bucket(aws *aws.AWS) (string, error) {
	accountID, err := aws.AccountID()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("mantil-%s-%s", aws.Region(), accountID), nil
}

func RuntimeResource(v ...string) string {
	r := "mantil"
	for _, n := range v {
		r = fmt.Sprintf("%s-%s", r, n)
	}
	return r
}

func SaveRuntimeConfig(aws *aws.AWS, rc *RuntimeConfig) error {
	bucket, err := Bucket(aws)
	if err != nil {
		return err
	}
	buf, err := json.Marshal(rc)
	if err != nil {
		return err
	}
	return aws.PutObjectToS3Bucket(bucket, s3RuntimeConfigKey, buf)
}

func LoadRuntimeConfig(aws *aws.AWS) (*RuntimeConfig, error) {
	bucket, err := Bucket(aws)
	if err != nil {
		return nil, err
	}
	buf, err := aws.GetObjectFromS3Bucket(bucket, s3RuntimeConfigKey)
	if err != nil {
		return nil, err
	}
	var c RuntimeConfig
	if err := json.Unmarshal(buf, &c); err != nil {
		return nil, err
	}
	return &c, nil
}

type RuntimeConfig struct {
	Version         string `json:"version"`
	FunctionsBucket string `json:"functionsBucket"`
	FunctionsPath   string `json:"functionsPath"`
}
