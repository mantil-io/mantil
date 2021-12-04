package test

import (
	"flag"
	"fmt"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/mantil-io/mantil/domain"
	"github.com/stretchr/testify/require"
)

func TestNodeResources(t *testing.T) {
	fs, err := domain.NewSingleDeveloperWorkspaceStore()
	require.NoError(t, err)
	testNodeResources(t, fs.Workspace())
}

func testNodeResources(t *testing.T, ws *domain.Workspace) {
	for _, node := range ws.Nodes {
		for _, rs := range node.Resources() {
			if rs.Type == domain.AwsResourceLambda {
				tags, err := awsCli.Lambda().Info(rs.AWSName)
				require.NoError(t, err)
				require.Equal(t, tags[domain.TagKey], node.ID)
				require.Equal(t, tags[domain.TagWorkspace], ws.ID)
				//t.Logf("lambda: %s, tags: %v", rs.AWSName, tags)
			}
		}
	}
}

func TestNodeCreateInDifferentRegions(t *testing.T) {
	// run this test only if explicitly called
	if !testNameIsInFlagRun(t) {
		t.Skip("this test can be only explicitly called by --run [name] flag")
	}
	if deadline, ok := t.Deadline(); !ok || time.Until(deadline) < 19*time.Minute {
		t.Skip("set timeout to 20 min or more for this test, default of 10min is probably not enough")
	}

	c := newClitest(t)

	// download mantil release binary
	archFilename := fmt.Sprintf("mantil_%s.tar.gz", target())
	c.Run("wget", "https://s3.eu-central-1.amazonaws.com/releases.mantil.io/latest/"+archFilename).Success()
	c.Run("tar", "xvfz", archFilename).Success()

	// show current mantil version
	mantilBin := c.GetWorkdir() + "/mantil"
	stdout := c.Run(mantilBin, "--version").Success().GetStdout()
	fmt.Printf("    %s", stdout)

	regions := []string{"ap-south-1", "ap-southeast-1", "ap-southeast-2", "ap-northeast-1", "eu-central-1", "eu-west-1", "eu-west-2", "us-east-1", "us-east-2", "us-west-2"}
	nodeName := "try"

	// run in each region
	for _, region := range regions {
		t.Run(region, func(t *testing.T) {
			c := newClitestWithWorkspaceCopy(t).
				Env("AWS_DEFAULT_REGION", region)

				// currently disabled, makes lots of ngs connections
				//t.Parallel()

			c.Run(mantilBin, "aws", "install", nodeName, "--aws-env").
				Contains(fmt.Sprintf("Mantil node %s created", nodeName))
			c.Run(mantilBin, "aws", "uninstall", nodeName, "--aws-env", "--force").
				Contains(fmt.Sprintf("Mantil node %s destroyed", nodeName))
		})
	}
}

func testNameIsInFlagRun(t *testing.T) bool {
	f := flag.Lookup("test.run")
	if f == nil {
		return false
	}
	if f.Value.String() == "" {
		return false
	}
	return strings.Contains(t.Name(), f.Value.String())
}

func target() string {
	tg := runtime.GOOS + "_" + runtime.GOARCH
	r := map[string]string{
		"darwin":  "Darwin",
		"linux":   "Linux",
		"windows": "Windows",
		"386":     "i386",
		"amd64":   "x86_64",
	}
	for k, v := range r {
		tg = strings.Replace(tg, k, v, 1)
	}
	return tg
}
