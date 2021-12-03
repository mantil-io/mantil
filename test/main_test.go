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

	"github.com/mantil-io/mantil/aws"
	"github.com/mantil-io/mantil/backend/secret"
	"github.com/mantil-io/mantil/domain"
	"github.com/mantil-io/mantil/kit/clitest"
	"github.com/stretchr/testify/require"
)

const defaultNodeName = "unit-test"
const awsAccountID = "052548195718"

var (
	cleanupWorkspace bool
	awsCli           *aws.AWS
	//commandOutputPath func(testName, name string, arg ...string) string
)

func setAwsEnv() {
	// org5 unit-test user
	os.Setenv("AWS_ACCESS_KEY_ID", "AKIAQYPA52WDEUQBVW2V")
	os.Setenv("AWS_SECRET_ACCESS_KEY", "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")
	os.Setenv("AWS_DEFAULT_REGION", "eu-central-1")
}

func TestMain(m *testing.M) {
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

	if !defaultNodeExists() {
		if err := createNewWorkspace(); err != nil {
			panic(err)
		}

		// create default node in the new workspace
		t := &testingT{}
		clitest.New(t).
			Run("mantil", "aws", "install", defaultNodeName, "--aws-env").Success()
		if len(t.errors) > 0 {
			return 3
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
			return 3
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

func createNewWorkspace() error {
	workspacePath, err := ioutil.TempDir("", "mantil-workspace-")
	if err != nil {
		return err
	}

	os.Setenv(domain.EnvWorkspacePath, workspacePath)
	fmt.Printf("setting workspace path to %s\n", workspacePath)

	// create and store activation token for this workspace
	jwt := secret.TokenForTests(domain.MachineID())
	return domain.StoreActivationTokenTo(jwt, workspacePath)
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

	// and access token
	jwt := secret.TokenForTests(domain.MachineID())
	err = domain.StoreActivationTokenTo(jwt, testDir)
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
