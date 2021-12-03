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
	_ = newCliRunner(t)

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
	if !testNameIsInFlagRun(t) {
		t.Skip("this test can be only explicitly called by --run [name] flag")
	}
	if deadline, ok := t.Deadline(); !ok || time.Until(deadline) < 20*time.Minute {
		t.Skip("set timeout to 20 min or more for this test, default of 10min is probably not enough")
	}

	r := newCliRunnerWithWorkspaceCopy(t)

	fn := fmt.Sprintf("mantil_%s.tar.gz", target())
	r.run("wget", "https://s3.eu-central-1.amazonaws.com/releases.mantil.io/latest/"+fn)
	r.run("tar", "xvfz", fn)

	mantilBin := r.testDir + "/mantil"
	c := r.run(mantilBin, "--version")
	fmt.Printf("    %s", c.Stdout())

	regions := []string{"ap-south-1", "ap-southeast-1", "ap-southeast-2", "ap-northeast-1", "eu-central-1", "eu-west-1", "eu-west-2", "us-east-1", "us-east-2", "us-west-2"}
	nodeName := "try"

	for _, region := range regions {
		t.Run(region, func(t *testing.T) {
			r := newCliRunnerWithWorkspaceCopy(t)
			r.SetEnv("AWS_DEFAULT_REGION", region)
			// currently disabled, makes lots of ngs connections
			//t.Parallel()

			r.Assert(fmt.Sprintf("Mantil node %s created", nodeName),
				mantilBin, "aws", "install", nodeName, "--aws-env")

			r.Assert(fmt.Sprintf("Mantil node %s destroyed", nodeName),
				mantilBin, "aws", "uninstall", nodeName, "--aws-env", "--force")
		})
	}
}

func testNameIsInFlagRun(t *testing.T) bool {
	f := flag.Lookup("test.run")
	if f == nil {
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
