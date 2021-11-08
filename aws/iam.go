package aws

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
)

type RoleNotExistsError struct {
	msg string
}

func (a *RoleNotExistsError) Error() string {
	return a.msg
}

func (a *AWS) Role(name string) (string, error) {
	gri := &iam.GetRoleInput{
		RoleName: aws.String(name),
	}
	r, err := a.iamClient.GetRole(context.Background(), gri)
	if err == nil {
		return aws.ToString(r.Role.Arn), nil
	}
	if strings.Contains(err.Error(), "NoSuchEntity") {
		return "", &RoleNotExistsError{msg: "role doesn't exist"}
	}
	return "", err
}

func (a *AWS) CreateRole(name, assumeRolePolicy, policy string) (string, error) {
	cri := &iam.CreateRoleInput{
		RoleName:                 aws.String(name),
		AssumeRolePolicyDocument: aws.String(assumeRolePolicy),
	}
	r, err := a.iamClient.CreateRole(context.Background(), cri)
	if err != nil {
		return "", fmt.Errorf("could not create role - %v", err)
	}

	rw := iam.NewRoleExistsWaiter(a.iamClient)
	if err := rw.Wait(context.Background(), &iam.GetRoleInput{
		RoleName: r.Role.RoleName,
	}, time.Minute); err != nil {
		return "", fmt.Errorf("error waiting for role - %v", err)
	}
	prpi := &iam.PutRolePolicyInput{
		PolicyDocument: aws.String(policy),
		PolicyName:     aws.String(name),
		RoleName:       aws.String(name),
	}
	_, err = a.iamClient.PutRolePolicy(context.Background(), prpi)
	if err != nil {
		return "", err
	}
	return aws.ToString(r.Role.Arn), nil
}

func iamRoleFromARN(arn string) (string, error) {
	resource, err := resourceFromARN(arn)
	if err != nil {
		return "", err
	}
	return strings.TrimPrefix(resource, "role/"), nil
}
