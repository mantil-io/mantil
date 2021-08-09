package aws

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	iamTypes "github.com/aws/aws-sdk-go-v2/service/iam/types"
)

func (a *AWS) CreateBootstrapRole(name, lambdaName string) (string, error) {
	r, err := a.createRole(name, bootstrapAssumeRolePolicy())
	if err != nil {
		return "", err
	}
	p, err := a.createPolicy(name, bootstrapLambdaPolicy(*r.RoleId, lambdaName))
	if err != nil {
		return "", err
	}
	if err := a.attachRolePolicy(*p.Arn, *r.RoleName); err != nil {
		return "", err
	}
	return *r.Arn, nil
}

func (a *AWS) createRole(name, policy string) (*iamTypes.Role, error) {
	cri := &iam.CreateRoleInput{
		RoleName:                 aws.String(name),
		AssumeRolePolicyDocument: aws.String(policy),
	}
	r, err := a.iamClient.CreateRole(context.Background(), cri)
	if err != nil {
		return nil, fmt.Errorf("could not create role - %v", err)
	}

	rw := iam.NewRoleExistsWaiter(a.iamClient)
	if err := rw.Wait(context.Background(), &iam.GetRoleInput{
		RoleName: r.Role.RoleName,
	}, time.Minute); err != nil {
		return nil, fmt.Errorf("error waiting for role - %v", err)
	}
	return r.Role, nil
}

func (a *AWS) createPolicy(name, policy string) (*iamTypes.Policy, error) {
	cpi := &iam.CreatePolicyInput{
		PolicyName:     aws.String(name),
		PolicyDocument: aws.String(policy),
	}
	p, err := a.iamClient.CreatePolicy(context.Background(), cpi)
	if err != nil {
		return nil, fmt.Errorf("could not create policy - %v", err)
	}

	pw := iam.NewPolicyExistsWaiter(a.iamClient)
	if err := pw.Wait(context.Background(), &iam.GetPolicyInput{
		PolicyArn: p.Policy.Arn,
	}, time.Minute); err != nil {
		return nil, fmt.Errorf("error waiting for policy - %v", err)
	}
	return p.Policy, nil
}

func (a *AWS) attachRolePolicy(policyArn, roleName string) error {
	arpi := &iam.AttachRolePolicyInput{
		PolicyArn: aws.String(policyArn),
		RoleName:  aws.String(roleName),
	}
	_, err := a.iamClient.AttachRolePolicy(context.Background(), arpi)
	if err != nil {
		return fmt.Errorf("could not attach policy - %v", err)
	}
	return nil
}

func bootstrapAssumeRolePolicy() string {
	return `{
		"Version": "2012-10-17",
		"Statement": [
			{
				"Action": "sts:AssumeRole",
				"Principal": {
					"Service": "lambda.amazonaws.com"
				},
				"Effect": "Allow"
			}
		]
	}`
}

func bootstrapLambdaPolicy(roleID, lambdaName string) string {
	return `{
		"Version": "2012-10-17",
		"Statement": [
			{
				"Effect": "Allow",
				"Resource": "*",
				"Action": "*"
			},
			{
				"Effect": "Deny",
				"Resource": "*",
				"Action": "*",
				"Condition": {
					"StringNotLike": {
						"aws:userid": "` + roleID + `:` + lambdaName + `"
					}
				}
			}
		]
	}`
}

func (a *AWS) DeleteRole(name string) error {
	larpi := &iam.ListAttachedRolePoliciesInput{
		RoleName: aws.String(name),
	}
	rsp, err := a.iamClient.ListAttachedRolePolicies(context.Background(), larpi)
	if err != nil {
		return fmt.Errorf("error listing role policies - %v", err)
	}
	for _, p := range rsp.AttachedPolicies {
		drpi := &iam.DetachRolePolicyInput{
			PolicyArn: p.PolicyArn,
			RoleName:  aws.String(name),
		}
		_, err := a.iamClient.DetachRolePolicy(context.Background(), drpi)
		if err != nil {
			return fmt.Errorf("error detaching role policy - %v", err)
		}
	}
	dri := &iam.DeleteRoleInput{
		RoleName: aws.String(name),
	}
	_, err = a.iamClient.DeleteRole(context.Background(), dri)
	if err != nil {
		return fmt.Errorf("error deleting role - %v", err)
	}
	return nil
}

func (a *AWS) DeletePolicy(name string) error {
	accountID, err := a.AccountID()
	if err != nil {
		return err
	}
	arn := fmt.Sprintf("arn:aws:iam::%s:policy/%s", accountID, name)
	dpi := &iam.DeletePolicyInput{
		PolicyArn: aws.String(arn),
	}
	_, err = a.iamClient.DeletePolicy(context.Background(), dpi)
	if err != nil {
		return fmt.Errorf("error deleting policy - %v", err)
	}
	return nil
}
