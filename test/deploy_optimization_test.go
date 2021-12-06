package test

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/mantil-io/mantil/domain"
	"github.com/stretchr/testify/require"
)

// kako izabrati mem limite za labda funkciju
// koje granice ima smisla testirati
// ref: https://www.sentiatechblog.com/aws-re-invent-2020-day-3-optimizing-lambda-cost-with-multi-threading?utm_source=reddit&utm_medium=social&utm_campaign=day3_lambda
func TestDeployOptimization(t *testing.T) {
	// run this test only if explicitly called
	if !testNameIsInFlagRun(t) {
		t.Skip("this test can be only explicitly called by --run [name] flag")
	}

	// protostavljam da ovdje imam projekt
	// da imam neki node koji je napravljen u
	dir := "/tmp/project1"
	os.Chdir(dir)
	c := newClitestWithWorkspaceCopy(t).Workdir(dir)

	var deployLambdaName, destroyLambdaName string
	// find lambda function name
	fs, err := domain.NewSingleDeveloperWorkspaceStore()
	require.NoError(t, err)
	node := fs.Workspace().Node(defaultNodeName)
	if node == nil {
		t.Fatalf("node %s not found", defaultNodeName)
	}

	for _, rs := range node.Resources() {
		if rs.Type == domain.AwsResourceLambda && rs.Name == "deploy" {
			deployLambdaName = rs.AWSName
		}
		if rs.Type == domain.AwsResourceLambda && rs.Name == "destroy" {
			destroyLambdaName = rs.AWSName
		}
	}

	type res struct {
		No      int
		Mem     int
		New     time.Duration
		Destroy time.Duration
	}
	var results []res

	save := func() {
		buf, _ := json.Marshal(results)
		err = ioutil.WriteFile("deploy_optimization.json", buf, 0600)
		require.NoError(t, err)
	}

	//steps := []int{512, 1024, 2048, 4096}
	//steps := []int{512, 2048}
	//steps := []int{1769, 3009, 5308, 7077, 8846}
	steps := []int{442, 885, 1769, 3009}
	for _, mem := range steps {

		for no := 1; no < 4; no++ {
			err := awsCli.Lambda().SetMemory(deployLambdaName, mem)
			require.NoError(t, err)
			err = awsCli.Lambda().SetMemory(destroyLambdaName, mem)
			require.NoError(t, err)

			stage := fmt.Sprintf("stage-%d-%d", mem, no)

			r := res{Mem: mem, No: no}
			r.New = c.Run("mantil", "stage", "new", stage, "--node", node.Name).Success().Duration()
			r.Destroy = c.Run("mantil", "stage", "destroy", stage, "--yes").Success().Duration()
			results = append(results, r)
		}

		save()
	}
	save()

	showOptimization(t)
}

func showOptimization(t *testing.T) {
	buf, err := ioutil.ReadFile("deploy_optimization.json")
	require.NoError(t, err)
	type res struct {
		No      int
		Mem     int
		New     time.Duration
		Destroy time.Duration
	}
	results := make([]res, 0)

	err = json.Unmarshal(buf, &results)
	require.NoError(t, err)

	cum := make(map[int]time.Duration)
	for _, r := range results {
		c := r.Destroy + r.New
		cum[r.Mem] += c
	}

	for k, v := range cum {
		fmt.Printf("%d %v\n", k, v)
	}
}
