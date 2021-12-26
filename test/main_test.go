package test

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/joho/godotenv"
	"github.com/mantil-io/mantil/kit/aws"
	"github.com/mantil-io/mantil/domain"
	"github.com/mantil-io/mantil/kit/clitest"
	"github.com/mantil-io/mantil/kit/testutil"
	"github.com/stretchr/testify/require"
)

const defaultNodeName = "unit-test"

var (
	cleanupWorkspace bool
	awsCli           *aws.AWS
	awsAccountID     string
)

func setAwsEnv() {
	if aws.InGithubAction() { // github actions sets session based tokens
		return
	}
	envFilename := "../../infrastructure/secrets/unit_test.env"
	if !testutil.FileExists(envFilename) {
		return
	}
	if err := godotenv.Load(envFilename); err != nil {
		panic("Error loading .env file")
	}
	awsAccountID = os.Getenv("AWS_ACCOUNT_ID")
}

func TestMain(m *testing.M) {
	os.Setenv(domain.EnvNoEvents, "1") // disable sending cli events
	flag.Parse()

	for _, f := range []func() int{setup, m.Run, teardown} {
		if code := f(); code != 0 {
			clitest.Show()
			os.Exit(code)
		}
	}
}

func setup() int {
	setAwsEnv()

	var err error
	awsCli, err = aws.New()
	if err != nil {
		panic(err)
	}

	// if node named unit-test exists in default workspace
	// and that node is in the unit-test account
	// it will be used instead of creating new
	// to speed up startup
	if !defaultNodeExists() {
		workspacePath, err := createNewWorkspace()
		if err != nil {
			panic(err)
		}
		os.Setenv(domain.EnvWorkspacePath, workspacePath)
		fmt.Printf("workspace path %s\n", workspacePath)

		// create default node in the new workspace
		t := &testingT{}
		clitest.New(t).
			Run("mantil", "aws", "install", defaultNodeName, "--aws-env").Success()
		if len(t.errors) > 0 {
			return 3 // m.Run uses code 1,2 so we use 3, 4
		}

		cleanupWorkspace = true
	} else {
		fmt.Printf("using node %s in default workspace\n", defaultNodeName)
	}
	return 0
}

func teardown() int {
	if cleanupWorkspace {
		fs, err := domain.NewSingleDeveloperWorkspaceStore()
		if err != nil {
			panic(err)
		}
		t := &testingT{}
		c := clitest.New(t)
		for _, n := range fs.Workspace().Nodes {
			c.Run("mantil", "aws", "uninstall", n.Name, "--aws-env", "--force").Success()
		}
		if len(t.errors) > 0 {
			return 4
		}
	}
	return 0
}

func defaultNodeExists() bool {
	fs, err := domain.NewSingleDeveloperWorkspaceStore()
	if err != nil {
		return false
	}
	if node := fs.Workspace().Node(defaultNodeName); node != nil {
		return node.AccountID == awsAccountID
	}
	return false
}

func createNewWorkspace() (string, error) {
	workspacePath, err := ioutil.TempDir("", "mantil-workspace-")
	if err != nil {
		return "", err
	}
	return workspacePath, nil
}

func newClitest(t *testing.T) *clitest.Env {
	t.Helper()
	testDir := createTestDir(t)
	return clitest.New(t).
		Workdir(testDir)
}

func newClitestWithWorkspaceCopy(t *testing.T) *clitest.Env {
	t.Helper()
	testDir := createTestDir(t)
	copyWorkspace(t, testDir)
	return clitest.New(t).
		Env(domain.EnvWorkspacePath, testDir).
		Workdir(testDir)
}

func createTestDir(t *testing.T) string {
	testName := strings.Replace(t.Name(), "/", "-", -1)
	testDir, err := ioutil.TempDir("", "mantil-"+testName+"-*")
	require.NoError(t, err)
	return testDir
}

func copyWorkspace(t *testing.T, testDir string) {
	dir, name, err := domain.WorkspacePathAndName()
	require.NoError(t, err)
	wsSourcePath := filepath.Join(dir, name)

	wsCopyPath := filepath.Join(testDir, name)
	err = clitest.Cp(wsSourcePath, wsCopyPath)
	require.NoError(t, err)

	require.NoError(t, err)
}

// implementing interface for clitest
// we are using that in setup/teardown where we don't have testing.T
type testingT struct {
	errors []error
}

func (t *testingT) Assert() {
	if len(t.errors) != 0 {
		panic(t.errors)
	}
}

func (t *testingT) Helper() {}
func (t *testingT) Errorf(format string, args ...interface{}) {
	err := fmt.Errorf(format, args...)
	t.errors = append(t.errors, err)
	fmt.Printf("Error %s\n", err)
}
func (t *testingT) Logf(format string, args ...interface{}) {
	if testing.Verbose() {
		fmt.Printf(format, args...)
		fmt.Println()
	}
}
func (t *testingT) Deadline() (deadline time.Time, ok bool) {
	return time.Now().Add(10 * time.Minute), true
}
func (t *testingT) Name() string { return "setup" }
