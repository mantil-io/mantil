package aws

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"mime"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/smithy-go"
)

func (a *AWS) S3() *S3 {
	return &S3{
		a:   a,
		cli: a.s3Client,
	}
}

type S3 struct {
	a   *AWS
	cli *s3.Client
}

func (a *S3) createBucket(name, region string) error {
	cbi := &s3.CreateBucketInput{
		Bucket: aws.String(name),
	}
	// us-east-1 is default region - adding location constraint results in invalid location constraint error
	if region != "us-east-1" {
		cbi.CreateBucketConfiguration = &types.CreateBucketConfiguration{
			LocationConstraint: types.BucketLocationConstraint(region),
		}
	}
	_, err := a.cli.CreateBucket(context.Background(), cbi)
	if err != nil {
		return fmt.Errorf("could not create bucket %s in %s - %v", name, region, err)
	}
	return nil
}

func (a *S3) CreateBucket(bucket string, tags map[string]string) error {
	exists, err := a.bucketExists(bucket)
	if err != nil {
		return fmt.Errorf("error checking if bucket %s exists - %w", bucket, err)
	}
	if exists {
		return nil
	}
	if err := a.createBucket(bucket, a.a.Region()); err != nil {
		return fmt.Errorf("error creating bucket %s - %w", bucket, err)
	}
	if err := a.tagBucket(bucket, tags); err != nil {
		return fmt.Errorf("error tagging bucket %s - %w", bucket, err)
	}
	return nil

}

func (a *S3) tagBucket(name string, tags map[string]string) error {
	pbti := &s3.PutBucketTaggingInput{
		Bucket: aws.String(name),
	}
	t := []types.Tag{}
	for k, v := range tags {
		t = append(t, types.Tag{
			Key:   aws.String(k),
			Value: aws.String(v),
		})
	}
	pbti.Tagging = &types.Tagging{
		TagSet: t,
	}
	_, err := a.cli.PutBucketTagging(context.Background(), pbti)
	if err != nil {
		return fmt.Errorf("could not tag bucket %s - %w", name, err)
	}
	return nil
}

func (a *S3) emptyS3Bucket(name string) error {
	return a.delete(name, "")
}

func (a *S3) delete(name string, prefix string) error {
	loi := &s3.ListObjectsV2Input{
		Bucket: aws.String(name),
	}

	// add / at the end of prefix if it it's not present
	// ListObjectsV2 ignores everything after last / in the string
	if prefix != "" && !strings.HasSuffix(prefix, "/") {
		prefix = fmt.Sprintf("%s/", prefix)
	}

	if prefix != "" {
		loi.Prefix = aws.String(prefix)
	}

	for {
		out, err := a.cli.ListObjectsV2(context.Background(), loi)
		if err != nil {
			return fmt.Errorf("emptying bucket failed - %v", err)
		}

		for _, object := range out.Contents {
			if err := a.deleteObject(name, *object.Key); err != nil {
				return fmt.Errorf("emptying bucket failed - %v", err)
			}
		}

		if out.IsTruncated {
			loi.ContinuationToken = out.ContinuationToken
		} else {
			break
		}
	}
	return nil
}

func (a *S3) DeleteBucket(name string) error {
	// all objects must be deleted before bucket can be deleted
	if err := a.emptyS3Bucket(name); err != nil {
		return fmt.Errorf("could not delete bucket %s - %v", name, err)
	}

	dbi := &s3.DeleteBucketInput{
		Bucket: aws.String(name),
	}

	_, err := a.cli.DeleteBucket(context.Background(), dbi)
	if err != nil {
		return fmt.Errorf("could not delete bucket %s - %v", name, err)
	}
	return nil
}

func (a *S3) bucketExists(name string) (bool, error) {
	hbi := &s3.HeadBucketInput{
		Bucket: aws.String(name),
	}

	_, err := a.cli.HeadBucket(context.Background(), hbi)
	if err != nil {
		var oe smithy.APIError
		if errors.As(err, &oe) {
			switch oe.ErrorCode() {
			case "Forbidden":
				return true, nil
			case "NotFound":
				return false, nil
			case "MovedPermanently":
				return true, nil
			default:
				return false, err
			}
		} else {
			return false, err
		}
	}
	return true, nil
}

func (a *S3) Put(bucket, key string, buf []byte) error {
	contentType := mime.TypeByExtension(filepath.Ext(key))
	poi := &s3.PutObjectInput{
		Bucket:      aws.String(bucket),
		Key:         aws.String(key),
		Body:        bytes.NewReader(buf),
		ContentType: aws.String(contentType),
	}
	_, err := a.cli.PutObject(context.Background(), poi)
	if err != nil {
		return fmt.Errorf("could not put key %s in bucket %s - %v", key, bucket, err)
	}
	return nil
}

func (a *S3) deleteObject(bucket, key string) error {
	doi := &s3.DeleteObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}

	_, err := a.cli.DeleteObject(context.Background(), doi)
	if err != nil {
		return fmt.Errorf("could not delete key %s in bucket %s - %v", bucket, key, err)
	}
	return nil
}

// func (a *AWS) S3PrefixExistsInBucket(name string, prefix string) (bool, error) {
// 	loi := &s3.ListObjectsV2Input{
// 		Bucket:  aws.String(name),
// 		MaxKeys: 1, // 1 object is enough to determine if prefix exists
// 		Prefix:  aws.String(prefix),
// 	}

// 	loo, err := a.s3Client.ListObjectsV2(context.Background(), loi)
// 	if err != nil {
// 		return false, err
// 	}
// 	return loo.KeyCount > 0, nil
// }

// func (a *AWS) GetObjectFromS3Bucket(bucket, key string) ([]byte, error) {
// 	goi := &s3.GetObjectInput{
// 		Bucket: aws.String(bucket),
// 		Key:    aws.String(key),
// 	}
// 	rsp, err := a.s3Client.GetObject(context.Background(), goi)
// 	var nsk *types.NoSuchKey
// 	if errors.As(err, &nsk) {
// 		err = ErrNotFound
// 	}
// 	if err != nil {
// 		return nil, fmt.Errorf("could not get key %s from bucket %s - %w", key, bucket, err)
// 	}
// 	defer rsp.Body.Close()
// 	buf, err := ioutil.ReadAll(rsp.Body)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return buf, nil
// }
