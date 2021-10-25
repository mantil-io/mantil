// Creates release buckets for all regions, enables versioning, adds necessary policy and adds replication rule for replicationPrefix in mainBucket
package main

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"strings"
	"text/template"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

var (
	replicationRole      = "mantil-releases-replication"
	mainBucket           = "mantil-releases"
	mainRegion           = "eu-central-1"
	replicationPrefix    = "v"
	regionBucketTemplate = "mantil-releases-%s"
)

func main() {
	c, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	regions, err := regions(c)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Regions discovered: %s", strings.Join(regions, ","))
	for _, r := range regions {
		log.Printf("Processing bucket for region %s.\n", r)
		c, err := config.LoadDefaultConfig(context.Background(), config.WithRegion(r))
		if err != nil {
			log.Fatal(err)
		}
		s3Client := s3.NewFromConfig(c)
		if err := processBucket(s3Client, r); err != nil {
			log.Fatal(err)
		}
	}
	iamClient := iam.NewFromConfig(c)
	log.Printf("Creating replication role.")
	replicationRoleArn, err := createReplicationRole(iamClient, regions)
	if err != nil {
		log.Fatal(err)
	}

	if err := addReplication(mainBucket, regions, replicationPrefix, replicationRoleArn); err != nil {
		log.Fatal(err)
	}
}

func regions(config aws.Config) ([]string, error) {
	ec2Client := ec2.NewFromConfig(config)
	dro, err := ec2Client.DescribeRegions(context.Background(), &ec2.DescribeRegionsInput{
		AllRegions: aws.Bool(true),
	})
	if err != nil {
		return nil, err
	}
	var regions []string
	for _, r := range dro.Regions {
		regions = append(regions, aws.ToString(r.RegionName))
	}
	return regions, nil
}

func createReplicationRole(client *iam.Client, regions []string) (string, error) {
	roleArn, err := createRole(client, replicationRole, replicationAssumeRolePolicy)
	if err != nil {
		return "", err
	}
	replicationPolicy, err := createReplicationPolicy(regions)
	if err != nil {
		return "", err
	}
	policyArn, err := createPolicy(client, replicationRole, replicationPolicy)
	if err != nil {
		return "", err
	}
	if err := attachRolePolicy(client, replicationRole, policyArn); err != nil {
		return "", err
	}
	return roleArn, nil
}

func createRole(client *iam.Client, name, policy string) (string, error) {
	cri := &iam.CreateRoleInput{
		RoleName:                 aws.String(name),
		AssumeRolePolicyDocument: aws.String(policy),
	}
	cro, err := client.CreateRole(context.Background(), cri)
	if err != nil {
		return "", err
	}
	waiter := iam.NewRoleExistsWaiter(client)
	if err := waiter.Wait(context.Background(), &iam.GetRoleInput{RoleName: aws.String(name)}, 2*time.Minute); err != nil {
		return "", err
	}
	return aws.ToString(cro.Role.Arn), nil
}

func createReplicationPolicy(regions []string) (string, error) {
	buckets := []string{mainBucket}
	for _, r := range regions {
		buckets = append(buckets, fmt.Sprintf(regionBucketTemplate, r))
	}
	data := struct {
		Buckets []string
	}{
		buckets,
	}
	tpl := template.Must(template.New("").Parse(replicationPolicyTemplate))
	buf := bytes.NewBuffer(nil)
	if err := tpl.Execute(buf, data); err != nil {
		return "", nil
	}
	return buf.String(), nil
}

func createPolicy(client *iam.Client, name, policy string) (string, error) {
	cpi := &iam.CreatePolicyInput{
		PolicyName:     aws.String(name),
		PolicyDocument: aws.String(policy),
	}
	cpo, err := client.CreatePolicy(context.Background(), cpi)
	if err != nil {
		return "", err
	}
	waiter := iam.NewPolicyExistsWaiter(client)
	if err := waiter.Wait(context.Background(), &iam.GetPolicyInput{PolicyArn: cpo.Policy.Arn}, 2*time.Minute); err != nil {
		return "", err
	}
	return aws.ToString(cpo.Policy.Arn), nil
}

func attachRolePolicy(client *iam.Client, role, policyArn string) error {
	arpi := &iam.AttachRolePolicyInput{
		RoleName:  aws.String(role),
		PolicyArn: aws.String(policyArn),
	}
	_, err := client.AttachRolePolicy(context.Background(), arpi)
	return err
}

func processBucket(client *s3.Client, region string) error {
	name := fmt.Sprintf(regionBucketTemplate, region)
	log.Println("Creating bucket...")
	if err := createBucket(client, name, region); err != nil {
		return err
	}
	log.Println("Enabling versioning...")
	if err := enableVersioning(client, name); err != nil {
		return err
	}
	log.Println("Deleting public access block...")
	if err := deletePublicAccessBlock(client, name); err != nil {
		return err
	}
	log.Println("Adding bucket policy...")
	if err := putBucketPolicy(client, name, fmt.Sprintf(bucketPolicyTemplate, name)); err != nil {
		return err
	}
	return nil
}

func createBucket(client *s3.Client, name, region string) error {
	cbi := &s3.CreateBucketInput{
		Bucket: aws.String(name),
	}
	if region != "us-east-1" {
		cbi.CreateBucketConfiguration = &types.CreateBucketConfiguration{
			LocationConstraint: types.BucketLocationConstraint(region),
		}
	}
	_, err := client.CreateBucket(context.Background(), cbi)
	return err
}

func enableVersioning(client *s3.Client, name string) error {
	pbvi := &s3.PutBucketVersioningInput{
		Bucket: aws.String(name),
		VersioningConfiguration: &types.VersioningConfiguration{
			Status: types.BucketVersioningStatusEnabled,
		},
	}
	_, err := client.PutBucketVersioning(context.Background(), pbvi)
	return err
}

func deletePublicAccessBlock(Client *s3.Client, name string) error {
	dpabi := &s3.DeletePublicAccessBlockInput{
		Bucket: aws.String(name),
	}
	_, err := Client.DeletePublicAccessBlock(context.Background(), dpabi)
	return err
}

func putBucketPolicy(client *s3.Client, name, policy string) error {
	pbpi := &s3.PutBucketPolicyInput{
		Bucket: aws.String(name),
		Policy: aws.String(policy),
	}
	_, err := client.PutBucketPolicy(context.Background(), pbpi)
	return err
}

func addReplication(name string, regions []string, filter, roleArn string) error {
	// request has to be made from the region of the main bucket
	c, err := config.LoadDefaultConfig(context.Background(), config.WithRegion(mainRegion))
	if err != nil {
		return err
	}
	client := s3.NewFromConfig(c)
	rules := []types.ReplicationRule{}
	for i, r := range regions {
		destination := fmt.Sprintf(regionBucketTemplate, r)
		rules = append(rules, types.ReplicationRule{
			Destination: &types.Destination{Bucket: aws.String(fmt.Sprintf("arn:aws:s3:::%s", destination))},
			Status:      types.ReplicationRuleStatusEnabled,
			Priority:    int32(i) + 1, // required
			DeleteMarkerReplication: &types.DeleteMarkerReplication{
				Status: types.DeleteMarkerReplicationStatusEnabled,
			}, // required
			Filter: &types.ReplicationRuleFilterMemberPrefix{Value: filter},
		})
	}
	pbri := &s3.PutBucketReplicationInput{
		Bucket: aws.String(name),
		ReplicationConfiguration: &types.ReplicationConfiguration{
			Role:  aws.String(roleArn),
			Rules: rules,
		},
	}
	_, err = client.PutBucketReplication(context.Background(), pbri)
	return err
}

var replicationAssumeRolePolicy = `{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": "s3.amazonaws.com"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}`

var replicationPolicyTemplate = `{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Action": [
                "s3:ListBucket",
                "s3:GetReplicationConfiguration",
                "s3:GetObjectVersionForReplication",
                "s3:GetObjectVersionAcl",
                "s3:GetObjectVersionTagging",
                "s3:GetObjectRetention",
                "s3:GetObjectLegalHold"
            ],
            "Effect": "Allow",
            "Resource": [
            {{ range $index, $element := .Buckets }}
                {{ if $index }},{{ end }}
                "arn:aws:s3:::{{.}}",
                "arn:aws:s3:::{{.}}/*"
            {{ end }}
            ]
        },
        {
            "Action": [
                "s3:ReplicateObject",
                "s3:ReplicateDelete",
                "s3:ReplicateTags",
                "s3:ObjectOwnerOverrideToBucketOwner"
            ],
            "Effect": "Allow",
            "Resource": [
            {{ range $index, $element := .Buckets }}
                {{ if $index }},{{ end }}
                "arn:aws:s3:::{{.}}/*"
            {{ end }}
            ]
        }
    ]
}`

var bucketPolicyTemplate = `{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Sid": "AllowPublicRead",
            "Effect": "Allow",
            "Principal": "*",
            "Action": "s3:GetObject",
            "Resource": "arn:aws:s3:::%s/*"
        }
    ]
}
`
