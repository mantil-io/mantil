package aws

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"regexp"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/ecr"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	lambdaTypes "github.com/aws/aws-sdk-go-v2/service/lambda/types"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/sts"
)

type AWS struct {
	config       aws.Config
	s3Client     *s3.Client
	lambdaClient *lambda.Client
	stsClient    *sts.Client
	ecrClient    *ecr.Client
	iamClient    *iam.Client
}

func New(accessKeyID, secretAccessKey, sessionToken string) (*AWS, error) {
	config, err := config.LoadDefaultConfig(context.TODO(),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			accessKeyID,
			secretAccessKey,
			sessionToken,
		)),
		config.WithRegion("eu-central-1"))
	if err != nil {
		return nil, fmt.Errorf("unable to load SDK configuration - %v", err)
	}

	if config.Region == "" {
		return nil, fmt.Errorf("default region is not specified - to specify a region either set the AWS_REGION environment variable or set the region through config file")
	}

	return clientFromConfig(config), nil
}

func NewFromProfile(profile string) (*AWS, error) {
	config, err := config.LoadDefaultConfig(
		context.TODO(),
		config.WithSharedConfigProfile(profile),
	)
	if err != nil {
		return nil, fmt.Errorf("unable to load SDK configuration - %v", err)
	}
	return clientFromConfig(config), nil
}

func ListProfiles() ([]string, error) {
	configFilePath := config.DefaultSharedConfigFilename()
	buf, err := ioutil.ReadFile(configFilePath)
	if err != nil {
		return nil, fmt.Errorf("could not read AWS credentials file - %v", err)
	}
	profileRegex := regexp.MustCompile(`^\[profile (.*?)\]`)
	var profiles []string
	for _, line := range strings.Split(string(buf), "\n") {
		if strings.HasPrefix(line, "[default]") {
			profiles = append(profiles, "default")
			continue
		}
		res := profileRegex.FindStringSubmatch(line)
		if len(res) > 0 {
			profiles = append(profiles, res[1])
		}
	}
	return profiles, nil
}

func clientFromConfig(config aws.Config) *AWS {
	return &AWS{
		config:       config,
		s3Client:     s3.NewFromConfig(config),
		lambdaClient: lambda.NewFromConfig(config),
		stsClient:    sts.NewFromConfig(config),
		ecrClient:    ecr.NewFromConfig(config),
		iamClient:    iam.NewFromConfig(config),
	}
}

func (a *AWS) Credentials() (aws.Credentials, error) {
	return a.config.Credentials.Retrieve(context.TODO())
}

func (a *AWS) PutObjectToS3Bucket(bucket, key string, object io.Reader) error {
	poi := &s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
		Body:   object,
	}

	_, err := a.s3Client.PutObject(context.TODO(), poi)
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
	rsp, err := a.s3Client.GetObject(context.TODO(), goi)
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

func (a *AWS) GetECRLogin() (string, string, error) {
	geto, err := a.ecrClient.GetAuthorizationToken(context.TODO(), &ecr.GetAuthorizationTokenInput{})
	if err != nil {
		return "", "", err
	}
	if len(geto.AuthorizationData) == 0 || geto.AuthorizationData[0].AuthorizationToken == nil {
		return "", "", fmt.Errorf("no authorization data returned for ECR")
	}

	at := *geto.AuthorizationData[0].AuthorizationToken
	dat, err := base64.StdEncoding.DecodeString(at)
	if err != nil {
		return "", "", err
	}

	login := strings.Split(string(dat), ":")
	if len(login) != 2 {
		return "", "", fmt.Errorf("login data wrong format")
	}
	return login[0], login[1], nil
}

func (a *AWS) CreateBootstrapRole(name, lambdaName string) (string, error) {
	iamClient := a.iamClient
	cri := &iam.CreateRoleInput{
		RoleName:                 aws.String(name),
		AssumeRolePolicyDocument: aws.String(bootstrapAssumeRolePolicy()),
	}
	r, err := iamClient.CreateRole(context.TODO(), cri)
	if err != nil {
		return "", fmt.Errorf("could not create role - %v", err)
	}
	rw := iam.NewRoleExistsWaiter(iamClient)
	if err := rw.Wait(context.TODO(), &iam.GetRoleInput{
		RoleName: r.Role.RoleName,
	}, time.Minute); err != nil {
		return "", fmt.Errorf("error waiting for role - %v", err)
	}
	cpi := &iam.CreatePolicyInput{
		PolicyName:     aws.String(name),
		PolicyDocument: aws.String(bootstrapLambdaPolicy(*r.Role.RoleId, lambdaName)),
	}
	p, err := iamClient.CreatePolicy(context.TODO(), cpi)
	if err != nil {
		return "", fmt.Errorf("could not create policy - %v", err)
	}
	pw := iam.NewPolicyExistsWaiter(iamClient)
	if err := pw.Wait(context.TODO(), &iam.GetPolicyInput{
		PolicyArn: p.Policy.Arn,
	}, time.Minute); err != nil {
		return "", fmt.Errorf("error waiting for policy - %v", err)
	}
	arpi := &iam.AttachRolePolicyInput{
		PolicyArn: p.Policy.Arn,
		RoleName:  r.Role.RoleName,
	}
	_, err = iamClient.AttachRolePolicy(context.TODO(), arpi)
	if err != nil {
		return "", fmt.Errorf("could not attach policy - %v", err)
	}
	return *r.Role.Arn, nil
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
	rsp, err := a.iamClient.ListAttachedRolePolicies(context.TODO(), larpi)
	if err != nil {
		return fmt.Errorf("error listing role policies - %v", err)
	}
	for _, p := range rsp.AttachedPolicies {
		drpi := &iam.DetachRolePolicyInput{
			PolicyArn: p.PolicyArn,
			RoleName:  aws.String(name),
		}
		_, err := a.iamClient.DetachRolePolicy(context.TODO(), drpi)
		if err != nil {
			return fmt.Errorf("error detaching role policy - %v", err)
		}
	}
	dri := &iam.DeleteRoleInput{
		RoleName: aws.String(name),
	}
	_, err = a.iamClient.DeleteRole(context.TODO(), dri)
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
	_, err = a.iamClient.DeletePolicy(context.TODO(), dpi)
	if err != nil {
		return fmt.Errorf("error deleting policy - %v", err)
	}
	return nil
}

func (a *AWS) CreateLambdaFunction(name, role, s3Bucket, s3Key string, layers []string) (string, error) {
	fc := &lambdaTypes.FunctionCode{
		S3Bucket: aws.String(s3Bucket),
		S3Key:    aws.String(s3Key),
	}
	cfi := &lambda.CreateFunctionInput{
		Code:         fc,
		FunctionName: aws.String(name),
		Handler:      aws.String("bootstrap"),
		Role:         aws.String(role),
		Runtime:      lambdaTypes.RuntimeProvidedal2,
		Timeout:      aws.Int32(60 * 15),
		MemorySize:   aws.Int32(512),
		Layers:       layers,
	}
	// lambda creation might fail if the corresponding execution role was just created so we retry until it succeeds
	retryInterval := time.Second
	retryAttempts := 60
	var rsp *lambda.CreateFunctionOutput
	var err error
	for retryAttempts > 0 {
		rsp, err = a.lambdaClient.CreateFunction(context.TODO(), cfi)
		if err == nil {
			break
		}
		if strings.Contains(err.Error(), "The role defined for the function cannot be assumed by Lambda") ||
			strings.Contains(err.Error(), "The provided execution role does not have permissions") {
			time.Sleep(retryInterval)
			retryAttempts--
			continue
		}
		if err != nil {
			return "", fmt.Errorf("could not create function - %v", err)
		}
	}
	w := lambda.NewFunctionActiveWaiter(a.lambdaClient)
	if err := w.Wait(context.TODO(), &lambda.GetFunctionConfigurationInput{
		FunctionName: rsp.FunctionArn,
	}, time.Minute); err != nil {
		return "", fmt.Errorf("error waiting for function - %v", err)
	}
	return *rsp.FunctionArn, nil
}

func (a *AWS) DeleteLambdaFunction(name string) error {
	dfi := &lambda.DeleteFunctionInput{
		FunctionName: aws.String(name),
	}
	_, err := a.lambdaClient.DeleteFunction(context.TODO(), dfi)
	if err != nil {
		return fmt.Errorf("error deleting lambda function - %v", err)
	}
	return nil
}

func (a *AWS) InvokeLambdaFunction(arn string, req, rsp, clientContext interface{}) error {
	payload, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("could not marshal request - %v", err)
	}
	lii := &lambda.InvokeInput{
		FunctionName: aws.String(arn),
		Payload:      payload,
	}
	if clientContext != nil {
		buf, err := json.Marshal(clientContext)
		if err != nil {
			return fmt.Errorf("could not marshal client context - %v", err)
		}
		b64Ctx := base64.StdEncoding.EncodeToString(buf)
		lii.ClientContext = aws.String(b64Ctx)
	}
	output, err := a.lambdaClient.Invoke(context.TODO(), lii)
	if err != nil {
		return fmt.Errorf("could not invoke lambda function - %v", err)
	}
	if rsp != nil {
		if err := json.Unmarshal(output.Payload, rsp); err != nil {
			return fmt.Errorf("could not unmarshal response - %v", err)
		}
	}
	return nil
}

func (a *AWS) AccountID() (string, error) {
	gcio, err := a.stsClient.GetCallerIdentity(context.TODO(), &sts.GetCallerIdentityInput{})
	if err != nil {
		return "", fmt.Errorf("could not get account ID - %v", err)
	}
	return aws.ToString(gcio.Account), nil
}

func (a *AWS) Region() string {
	return a.config.Region
}
