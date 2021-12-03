package test

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"testing"

	"github.com/mantil-io/mantil/aws"
	"github.com/mantil-io/mantil/backend/secret"
	"github.com/mantil-io/mantil/domain"
	"github.com/mantil-io/mantil/kit/testcli"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const defaultNodeName = "unit-test"

var (
	cleanupWorkspace  bool
	awsCli            *aws.AWS
	commandOutputPath func(testName, name string, arg ...string) string
)

func TestMain(m *testing.M) {
	setup()
	code := m.Run()
	tCode := teardown()
	if code != 0 {
		os.Exit(code)
	}
	os.Exit(tCode)
}

func setup() {
	setAwsEnv()

	var err error
	awsCli, err = aws.New()
	if err != nil {
		panic(err)
	}

	createCommandsOutputDir()

	if !defaultNodeExists() {
		if err := createNewWorkspace(); err != nil {
			panic(err)
		}
		cleanupWorkspace = true
	} else {
		fmt.Printf("using node %s in default workspace\n", defaultNodeName)
	}
}

func teardown() int {
	if cleanupWorkspace {
		fs, err := domain.NewSingleDeveloperWorkspaceStore()
		if err != nil {
			panic(err)
		}
		r := newTeardownCliRunner()
		for _, n := range fs.Workspace().Nodes {
			deleteNode(r, n.Name)
		}
		return r.ExitCode()
	}
	return 0
}

func createCommandsOutputDir() {
	outputDir, err := ioutil.TempDir("/tmp", "test-output-*")
	if err != nil {
		panic(err)
	}

	var cnt int
	var mu sync.Mutex

	commandOutputPath = func(testName, name string, arg ...string) string {
		mu.Lock()
		cnt++
		c := cnt
		mu.Unlock()

		fn := fmt.Sprintf("%s-%03d-%s-%s", testName, c, name, strings.Join(arg, "_"))

		// replace all non-alphanumeric characters
		reg, err := regexp.Compile("[^A-Za-z0-9-_]+")
		if err != nil {
			log.Fatal(err)
		}
		fn = reg.ReplaceAllString(fn, "_")
		return filepath.Join(outputDir, fn)
	}

	fmt.Printf("commands output dir %s\n", outputDir)
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
	workspacePath, err := ioutil.TempDir("/tmp", "mantil-workspace-")
	if err != nil {
		return err
	}

	os.Setenv(domain.EnvWorkspacePath, workspacePath)
	fmt.Printf("setting workspace path to %s\n", workspacePath)

	// create and store activation token for this workspace
	jwt := secret.TokenForTests(domain.MachineID())
	return domain.StoreActivationToken(jwt)
}

type cliRunner struct {
	t       *testing.T
	fail    bool
	env     []string
	testDir string
	workDir string
}

func newTeardownCliRunner() *cliRunner {
	return &cliRunner{}
}

func newCliRunner(t *testing.T) *cliRunner {
	t.Helper()
	r := &cliRunner{
		t: t,
	}
	r.ensureDefaultNodeExists()
	return r
}

func newCliRunnerWithWorkspaceCopy(t *testing.T) *cliRunner {
	t.Helper()
	r := newCliRunner(t)

	r.copyWorkspace()
	r.workDir = r.TestDir()

	jwt := secret.TokenForTests(domain.MachineID())
	err := domain.StoreActivationTokenTo(jwt, r.workDir)
	require.NoError(t, err)

	// use current environment with replaced EnvWorkspacePath variable
	// for all shell commands
	var env []string
	for _, v := range os.Environ() {
		if !strings.Contains(v, domain.EnvWorkspacePath) {
			env = append(env, v)
		}
	}
	env = append(env, fmt.Sprintf("%s=%s", domain.EnvWorkspacePath, r.TestDir()))
	r.env = env
	return r
}

func (r *cliRunner) SetWorkdir(dir string) string {
	r.workDir = filepath.Join(r.TestDir(), dir)
	return r.workDir
}

func (r *cliRunner) TestDir() string {
	if r.testDir == "" {
		testName := strings.Replace(r.t.Name(), "/", "-", -1)
		testDir, err := ioutil.TempDir("/tmp", "mantil-"+testName+"-*")
		require.NoError(r.t, err)
		r.testDir = testDir
	}
	return r.testDir
}

func (r *cliRunner) copyWorkspace() {
	dir, name, err := domain.WorkspacePathAndName()
	require.NoError(r.t, err)
	wsSourcePath := filepath.Join(dir, name)

	wsCopyPath := filepath.Join(r.TestDir(), name)
	err = cp(wsSourcePath, wsCopyPath)
	require.NoError(r.t, err)
}

func (r *cliRunner) ExitCode() int {
	if r.t != nil {
		r.t.Helper()
	}
	if r.fail {
		return 1
	}
	return 0
}

func (r *cliRunner) run(name string, arg ...string) *testcli.Cmd {
	testName := ""
	if r.t != nil {
		r.t.Helper()
		testName = r.t.Name()
	}
	cmdStr := fmt.Sprintf("$ %s %s", name, strings.Join(arg, " "))
	r.logf("%s", cmdStr)

	cmd := testcli.Command(r.t, name, arg...)
	if len(r.env) != 0 {
		cmd.SetEnv(r.env)
	}
	if r.workDir != "" {
		cmd.SetWorkdir(r.workDir)
	}
	cmd.SetOutputPath(commandOutputPath(testName, name, arg...))
	duration := cmd.Run()

	r.logf("%s done in %v", cmdStr, duration)
	return cmd
}

func (r *cliRunner) logf(format string, arg ...interface{}) {
	if r.t != nil {
		r.t.Helper()
	}
	if r.t != nil {
		if !testing.Verbose() {
			return
		}
		format = fmt.Sprintf("    %-20s ", r.t.Name()) + format
	}
	format = format + "\n"
	fmt.Printf(format, arg...)
}

func (r *cliRunner) Assert(stdoutContains string, name string, arg ...string) *testcli.Cmd {
	if r.t != nil {
		r.t.Helper()
	}
	c := r.run(name, arg...)
	cmdStr := fmt.Sprintf("%s %s", name, strings.Join(arg, " "))
	if !c.Success() {
		desc := fmt.Sprintf("command %s FAILED\n", cmdStr)
		if r.t != nil {
			assert.Fail(r.t, desc)
		} else {
			fmt.Println(desc)
		}
		r.fail = true
	}
	r.StdoutContains(c, stdoutContains)
	return c
}

func (r *cliRunner) StdoutContains(c *testcli.Cmd, str string) {
	if str != "" && !c.StdoutContains(str) {
		desc := fmt.Sprintf("FAIL command %s output should contain %s", c.CmdStr(), str)
		if r.t != nil {
			assert.Fail(r.t, desc)
		} else {
			fmt.Println(desc)
		}
		r.fail = true
	}
}

func (r *cliRunner) ensureDefaultNodeExists() {
	if r.t == nil { // don't run during setup/teardown
		return
	}
	if !defaultNodeExists() {
		createNode(r, defaultNodeName)
	}
}

func (r *cliRunner) CpToProject(from, to string) {
	err := cp(from, filepath.Join(r.workDir, to))
	assert.NoError(r.t, err)
}

var globalMu sync.Mutex

func (r *cliRunner) WithWorkdir(cb func()) {
	globalMu.Lock()
	defer globalMu.Unlock()
	cwd, err := os.Getwd()
	os.Chdir(r.workDir)
	cb()
	if err == nil {
		os.Chdir(cwd)
	}
}

func (r *cliRunner) SetEnv(k, v string) {
	n, _ := r.envWithout(k)
	r.env = n
	r.env = append(r.env, fmt.Sprintf("%s=%s", k, v))
}

func (r *cliRunner) AddPath(v string) {
	n, p := r.envWithout("PATH")
	p = v + ":" + p
	r.env = n
	r.env = append(r.env, fmt.Sprintf("PATH=%s", p))
}

func (r *cliRunner) envWithout(key string) ([]string, string) {
	var n []string
	var val string
	key = key + "="
	for _, e := range r.env {
		if strings.HasPrefix(e, key) {
			val = strings.Split(e, "=")[1]
			continue
		}
		n = append(n, e)
	}
	return n, val
}

const awsAccountID = "052548195718"

func setAwsEnv() {
	// org5 unit-test user
	os.Setenv("AWS_ACCESS_KEY_ID", "AKIAQYPA52WDEUQBVW2V")
	os.Setenv("AWS_SECRET_ACCESS_KEY", "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")
	os.Setenv("AWS_DEFAULT_REGION", "eu-central-1")
}

func createNode(r *cliRunner, name string) {
	r.Assert(fmt.Sprintf("Mantil node %s created", name),
		"mantil", "aws", "install", name, "--aws-env")
}
func deleteNode(r *cliRunner, name string) {
	r.Assert(fmt.Sprintf("Mantil node %s destroyed", name),
		"mantil", "aws", "uninstall", name, "--aws-env", "--force")
}

// Copy the src file to dst. Any existing file will be overwritten and will not
// copy file attributes.
func cp(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	if err != nil {
		return err
	}
	return out.Close()
}
