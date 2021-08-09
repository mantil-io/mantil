package aws

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

func (a *AWS) PutObjectToS3Bucket(bucket, key string, object io.Reader) error {
	poi := &s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
		Body:   object,
	}
	_, err := a.s3Client.PutObject(context.Background(), poi)
	if err != nil {
		return fmt.Errorf("could not put key %s in bucket %s - %v", bucket, key, err)
	}
	return nil
}

func (a *AWS) GetObjectFromS3Bucket(bucket, key string, o interface{}) error {
	goi := &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}
	rsp, err := a.s3Client.GetObject(context.Background(), goi)
	if err != nil {
		return fmt.Errorf("could not get key %s from bucket %s - %v", bucket, key, err)
	}
	defer rsp.Body.Close()

	decoder := json.NewDecoder(rsp.Body)
	if err := decoder.Decode(&o); err != nil {
		return err
	}
	return nil
}
