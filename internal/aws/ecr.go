package aws

import (
	"context"
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/ecr"
)

func (a *AWS) GetECRLogin() (string, string, error) {
	geto, err := a.ecrClient.GetAuthorizationToken(context.Background(), &ecr.GetAuthorizationTokenInput{})
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
