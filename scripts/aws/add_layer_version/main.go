// Creates layerName layer in all available regions from layerZipFile with action permission.
package main

import (
	"context"
	"io/ioutil"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/aws-sdk-go-v2/service/lambda/types"
)

var (
	layerName    = "terraform-layer"
	layerZipFile = "terraform-layer.zip"
	action       = "lambda:GetLayerVersion"
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
	for _, r := range regions {
		log.Printf("Adding layer %s for region %s.\n", layerName, r)
		c, err := config.LoadDefaultConfig(context.Background(), config.WithRegion(r))
		if err != nil {
			log.Fatal(err)
		}
		lambdaClient := lambda.NewFromConfig(c)
		if err := processLayer(lambdaClient, layerName, layerZipFile, action); err != nil {
			log.Fatal(err)
		}
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

func processLayer(client *lambda.Client, name, file, action string) error {
	content, err := ioutil.ReadFile(file)
	if err != nil {
		return err
	}
	log.Println("Publishing layer...")
	version, err := publishLayerVersion(client, name, content)
	if err != nil {
		return err
	}
	log.Printf("Adding permission %s for version %d.\n", action, version)
	if err := addLayerVersionPermission(client, name, version, action); err != nil {
		return err
	}
	return nil
}

func publishLayerVersion(client *lambda.Client, name string, content []byte) (int64, error) {
	plvi := &lambda.PublishLayerVersionInput{
		LayerName:          aws.String(name),
		CompatibleRuntimes: []types.Runtime{types.RuntimeProvidedal2},
		Content: &types.LayerVersionContentInput{
			ZipFile: content,
		},
	}
	plvo, err := client.PublishLayerVersion(context.Background(), plvi)
	if err != nil {
		return 0, err
	}
	return plvo.Version, nil
}

func addLayerVersionPermission(client *lambda.Client, name string, version int64, action string) error {
	alvpi := &lambda.AddLayerVersionPermissionInput{
		LayerName:     aws.String(name),
		VersionNumber: version,
		Principal:     aws.String("*"),
		Action:        aws.String(action),
		StatementId:   aws.String(strings.Split(action, ":")[1]),
	}
	_, err := client.AddLayerVersionPermission(context.Background(), alvpi)
	return err
}
