package domain_test

import (
	"fmt"
	"testing"

	. "github.com/mantil-io/mantil/domain"
	"github.com/stretchr/testify/require"
)

func TestStageResourceNaming(t *testing.T) {
	stage := testStage(t)

	require.Equal(t, "functions/my-project/my-stage", stage.FunctionsBucketPrefix())
	require.Equal(t, "state/my-project/my-stage", stage.StateBucketPrefix())
	require.Equal(t, []string{stage.FunctionsBucketPrefix(), stage.StateBucketPrefix()}, stage.BucketPrefixes())
	require.Equal(t, "my-project-my-stage", stage.LogGroupsPrefix())
	require.Equal(t, "my-project-my-stage-func1-abcdefg", stage.Functions[0].LambdaName())
	require.Equal(t, "my-project-my-stage-%s-abcdefg", stage.ResourceNamingTemplate())
}

func TestStageResourceTags(t *testing.T) {
	stage := testStage(t)
	tags := stage.ResourceTags()
	require.NotEmpty(t, tags)

	require.Equal(t, "abcdefg", tags[TagKey])
	require.Equal(t, "my-project", tags[TagProjectName])
	require.Equal(t, "my-stage", tags[TagStageName])
}

func TestStageAuthToken(t *testing.T) {
	stage := testStage(t)
	token, err := stage.AuthToken()

	require.NoError(t, err)
	require.NotEmpty(t, token)
}

func TestStageAuthEnv(t *testing.T) {
	stage := testStage(t)
	ae := stage.AuthEnv()

	require.NotEmpty(t, ae[EnvPublicKey])
}

func TestStageSetPublicBucket(t *testing.T) {
	stage := testStage(t)
	stage.SetPublicBucket("bucket")

	require.NotNil(t, stage.Public)
	require.Equal(t, "bucket", stage.Public.Bucket)
}

func TestStageSetEndpoints(t *testing.T) {
	stage := testStage(t)
	stage.SetEndpoints("rest", "ws")

	require.Equal(t, "rest", stage.Endpoints.Rest)
	require.Equal(t, "ws", stage.Endpoints.Ws)
}

func TestStageSetLastDeployment(t *testing.T) {
	stage := testStage(t)
	stage.SetLastDeployment()

	require.NotNil(t, stage.LastDeployment)
	require.Equal(t, stage.Node().Version, stage.LastDeployment.Version)
	require.NotEmpty(t, stage.LastDeployment.Timestamp)
}

func TestStageWsEnv(t *testing.T) {
	stage := testStage(t)
	wsEnv := stage.WsEnv()

	require.NotEmpty(t, wsEnv[EnvMantilConfig])
}

func TestStageFindFunction(t *testing.T) {
	stage := testStage(t)

	f := stage.FindFunction("func1")
	require.NotNil(t, f)

	f = stage.FindFunction("non-existent-func")
	require.Nil(t, f)
}

func TestStageRestEndpoint(t *testing.T) {
	stage := testStage(t)
	require.Empty(t, stage.RestEndpoint())

	stage.Endpoints = &StageEndpoints{
		Rest: "rest",
	}

	require.Equal(t, stage.Endpoints.Rest, stage.RestEndpoint())

	stage.CustomDomain = CustomDomain{
		DomainName: "domain",
	}
	require.Equal(t, fmt.Sprintf("https://%s", stage.CustomDomain.DomainName), stage.RestEndpoint())

	stage.CustomDomain.HttpSubdomain = "subdomain"
	require.Equal(t, fmt.Sprintf("https://%s.%s", stage.CustomDomain.HttpSubdomain, stage.CustomDomain.DomainName), stage.RestEndpoint())
}

func TestStageWsEndpoint(t *testing.T) {
	stage := testStage(t)
	require.Empty(t, stage.WsEndpoint())

	stage.Endpoints = &StageEndpoints{
		Ws: "ws",
	}

	require.Equal(t, fmt.Sprintf("%s/$default", stage.Endpoints.Ws), stage.WsEndpoint())

	stage.CustomDomain = CustomDomain{
		DomainName:  "domain",
		WsSubdomain: "subdomain",
	}
	require.Equal(t, fmt.Sprintf("wss://%s.%s", stage.CustomDomain.WsSubdomain, stage.CustomDomain.DomainName), stage.WsEndpoint())
}

func TestStagePublicBucketName(t *testing.T) {
	stage := testStage(t)
	require.Equal(t, "my-project-my-stage-public-abcdefg", stage.PublicBucketName())
}

func TestStagePublicEnv(t *testing.T) {
	stage := testStage(t)
	stage.Endpoints = &StageEndpoints{
		Rest: "rest",
		Ws:   "ws",
	}

	publicEnv, err := stage.PublicEnv()
	require.NoError(t, err)
	require.Equal(t, `var mantilEnv = {
	endpoints: {
		rest: 'rest',
		ws: 'ws/$default',
	},
};
`, string(publicEnv))
}

func TestStageCliStage(t *testing.T) {
	var stage *Stage
	require.Nil(t, stage.AsCliStage())

	stage = testStage(t)
	require.NotNil(t, stage.AsCliStage())
	require.Equal(t, stage.Name, stage.AsCliStage().Name)
}

func TestStageResources(t *testing.T) {
	stage := testStage(t)
	resources := stage.Resources()
	require.NotEmpty(t, resources)
	require.Len(t, resources, 7)

	stage.Public = &Public{}
	require.Len(t, stage.Resources(), 8)
}

func TestAwsResourceLogGroup(t *testing.T) {
	r := AwsResource{}
	require.Empty(t, r.LogGroup())

	r.Type = AwsResourceLambda
	require.Contains(t, r.LogGroup(), "/aws/lambda/")

	r.Type = AwsResourceAPIGateway
	require.Contains(t, r.LogGroup(), "/aws/vendedlogs/")
}

func testStage(t *testing.T) *Stage {
	workspace := Workspace{
		ID: "my-workspace-id",
		Nodes: []*Node{
			{
				Name:    "node1",
				ID:      "abcdefg",
				Version: "version",
			},
		},
	}
	stage := Stage{
		Name:     "my-stage",
		Default:  true,
		NodeName: "node1",
		Keys: StageKeys{
			Public:  "Y9ji3q73IEorB3ErJkQW_sFfvlUUQfR8j7ytGnW8LAk",
			Private: "_29hBmjkwUlIfL5uBjuZb6GQjfUwLsU_JnSvppB4G9hj2OLervcgSisHcSsmRBb-wV--VRRB9HyPvK0adbwsCQ",
		},
		Functions: []*Function{
			{
				Name: "func1",
			},
		},
	}
	project := Project{
		Name: "my-project",
		Stages: []*Stage{
			&stage,
		},
	}
	Factory(&workspace, &project, nil)
	return &stage
}
