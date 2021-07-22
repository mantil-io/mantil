package docker

import (
	"bufio"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"

	"github.com/atoz-technology/mantil-cli/internal/aws"
	"github.com/atoz-technology/mantil-cli/internal/mantil"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/archive"
)

const (
	ECRRegistry = "477361877445.dkr.ecr.eu-central-1.amazonaws.com"
)

func ProcessFunctionImage(f mantil.Function, repo, dir string) (string, error) {
	dc, err := client.NewClientWithOpts(client.WithAPIVersionNegotiation())
	if err != nil {
		return "", fmt.Errorf("unable to initialise docker client - %v", err)
	}
	image := fmt.Sprintf("%s/%s:%s-%s", ECRRegistry, repo, f.Name, f.Hash)

	if err := buildFunctionImage(dc, image, dir); err != nil {
		return "", err
	}

	if err := pushFunctionImage(dc, image); err != nil {
		return "", err
	}
	return image, nil
}

func buildFunctionImage(dc *client.Client, tag, dir string) error {
	tar, err := archive.TarWithOptions(dir, &archive.TarOptions{})
	if err != nil {
		return err
	}

	ibo := types.ImageBuildOptions{
		Dockerfile: "Dockerfile",
		Tags:       []string{tag},
	}

	res, err := dc.ImageBuild(context.TODO(), tar, ibo)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	return printAndCheckError(res.Body)
}

func pushFunctionImage(dc *client.Client, tag string) error {
	a, err := aws.New()
	if err != nil {
		return fmt.Errorf("unable to connect to aws - %v", err)
	}

	username, password, err := a.GetECRLogin()
	if err != nil {
		return fmt.Errorf("unable to get ECR login - %v", err)
	}

	authConfig := types.AuthConfig{
		Username: username,
		Password: password,
	}

	authConfigB, _ := json.Marshal(authConfig)
	authConfigE := base64.URLEncoding.EncodeToString(authConfigB)

	ipo := types.ImagePushOptions{RegistryAuth: authConfigE}

	res, err := dc.ImagePush(context.TODO(), tag, ipo)
	if err != nil {
		return fmt.Errorf("unable to push docker image to ECR - %v", err)
	}
	defer res.Close()
	return printAndCheckError(res)
}

func printAndCheckError(r io.Reader) error {
	var lastLine string

	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		lastLine = scanner.Text()
		log.Println(scanner.Text())
	}

	errLine := &ErrorLine{}
	if err := json.Unmarshal([]byte(lastLine), errLine); err != nil {
		return fmt.Errorf("unable to check if there was error while buidling image")
	}
	if errLine.Error != "" {
		return fmt.Errorf("error while building image - %s - %s", errLine.Error, errLine.ErrorDetail.Message)
	}
	return nil
}

type ErrorLine struct {
	Error       string      `json:"error"`
	ErrorDetail ErrorDetail `json:"errorDetail"`
}

type ErrorDetail struct {
	Message string `json:"message"`
}
