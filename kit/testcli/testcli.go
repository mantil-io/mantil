// CLI testing package for the Go language.
//
// Developing a command line application? Wanna be able to test your app from the
// outside? If the answer is Yes to at least one of the questions, keep reading.
//
// `testcli` is a wrapper around os/exec to test CLI apps in Go lang,
// minimalistic, so you can do your tests with `testing` or any other testing
// framework.
package testcli

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"sync"
	"testing"
	"time"
)

type output struct {
	content string
	mu      *sync.Mutex
}

// Cmd is typically constructed through the Command() call and provides state
// to the execution engine.
type Cmd struct {
	cmd        *exec.Cmd
	env        []string
	exitError  error
	status     string
	stdout     *output
	stderr     *output
	stdin      io.Reader
	t          *testing.T
	outputPath string
	cmdStr     string
}

// ErrUninitializedCmd is returned when members are accessed before a run, that
// can only be used after a command has been run.
var ErrUninitializedCmd = errors.New("You need to run this command first")

// ErrCmdNotFinished is returned when members are accessed before or during a run,
// that can only be used after a command has finished executing.
var ErrCmdNotFinished = errors.New("Command is still executing")

const (
	// INITIALIZED represents the state of Command before it's started with Run() or Start()
	initialized = "initialized"
	// RUNNING represents the state of Command while it's running
	running = "running"
	// FINISHED represents the state of Command after it has exited successfully or not
	finished = "finished"
)

var pkgCmd *Cmd

// Command constructs a *Cmd. It is passed the command name and arguments.
func Command(t *testing.T, name string, arg ...string) *Cmd {
	cmdStr := fmt.Sprintf("$ %s %s", name, strings.Join(arg, " "))
	return &Cmd{
		cmd:    exec.Command(name, arg...),
		t:      t,
		status: initialized,
		stdout: &output{mu: &sync.Mutex{}},
		stderr: &output{mu: &sync.Mutex{}},
		cmdStr: cmdStr,
	}
}

func (c *Cmd) CmdStr() string {
	return c.cmdStr
}

func (c *Cmd) SetWorkdir(dir string) {
	c.cmd.Dir = dir
}

func (c *Cmd) validateIsFinished() {
	if c.t != nil {
		c.t.Helper()
	}
	if c.status != finished {
		c.fatal(ErrCmdNotFinished)
	}
}

func (c *Cmd) fatal(err error) {
	if c.t == nil {
		panic(ErrCmdNotFinished)
	}
	c.t.Fatal(ErrCmdNotFinished)
}

func (c *Cmd) validateHasStarted() {
	if c.t != nil {
		c.t.Helper()
	}
	// After calling Start() status can either be running or finished
	if !(c.status == running || c.status == finished) {
		c.fatal(ErrUninitializedCmd)
	}
}

// SetEnv overwrites the environment with the provided one. Otherwise, the
// parent environment will be supplied.
func (c *Cmd) SetEnv(env []string) {
	c.env = env
}

// SetStdin sets the stdin stream. It makes no attempt to determine if the
// command accepts anything over stdin.
func (c *Cmd) SetStdin(stdin io.Reader) {
	c.stdin = stdin
}

func (c *Cmd) SetOutputPath(path string) {
	c.outputPath = path
}

// Run runs the command.
func (c *Cmd) Run() time.Duration {
	start := time.Now()
	if c.t != nil {
		c.t.Helper()
	}
	if c.stdin != nil {
		c.cmd.Stdin = c.stdin
	}

	if c.env != nil {
		c.cmd.Env = c.env
	} else {
		c.cmd.Env = os.Environ()
	}

	var outBuf bytes.Buffer
	c.cmd.Stdout = &outBuf

	var errBuf bytes.Buffer
	c.cmd.Stderr = &errBuf

	if err := c.cmd.Run(); err != nil {
		c.exitError = err
	}
	c.stdout.content = string(outBuf.Bytes())
	c.stderr.content = string(errBuf.Bytes())
	c.status = finished

	c.saveOutput()
	return time.Now().Sub(start).Round(time.Millisecond)
}

func (c *Cmd) saveOutput() {
	if c.outputPath == "" {
		return
	}
	if se := c.Stderr(); se != "" {
		fn := c.outputPath + ".stderr"
		err := ioutil.WriteFile(fn, []byte(se), 0644)
		if c.t != nil && err != nil {
			c.t.Logf("failed to save stdout output to %s, error: %s", fn, err)
		}
	}
	if so := c.Stdout(); so != "" {
		fn := c.outputPath + ".stdout"
		err := ioutil.WriteFile(fn, []byte(so), 0644)
		if c.t != nil && err != nil {
			c.t.Logf("failed to save stdout output to %s, error: %s", fn, err)
		}
	}
}

// Start starts the command without waiting for it to complete
func (c *Cmd) Start() {
	if c.t != nil {
		c.t.Helper()
	}
	if c.stdin != nil {
		c.cmd.Stdin = c.stdin
	}

	if c.env != nil {
		c.cmd.Env = c.env
	} else {
		c.cmd.Env = os.Environ()
	}

	stdoutPipe, err := c.cmd.StdoutPipe()
	if err != nil {
		c.fatal(err)
	}

	stderrPipe, err := c.cmd.StderrPipe()
	if err != nil {
		c.fatal(err)
	}

	if err := c.cmd.Start(); err != nil {
		c.exitError = err
	}

	go func() {
		scanner := bufio.NewScanner(stdoutPipe)
		for scanner.Scan() {
			c.stdout.mu.Lock()
			c.stdout.content += scanner.Text() + "\n"
			c.stdout.mu.Unlock()
		}
		if err := scanner.Err(); err != nil {
			c.fatal(err)
		}
	}()

	go func() {
		scanner := bufio.NewScanner(stderrPipe)
		for scanner.Scan() {
			c.stderr.mu.Lock()
			c.stderr.content += scanner.Text() + "\n"
			c.stderr.mu.Unlock()
		}
		if err := scanner.Err(); err != nil {
			c.fatal(err)
		}
	}()
	c.status = running
}

// Wait waits for a command started with Start() to exit.
// Will fail if called before Start() or Run()
func (c *Cmd) Wait() {
	if c.t != nil {
		c.t.Helper()
	}
	c.validateHasStarted()
	if err := c.cmd.Wait(); err != nil {
		c.exitError = err
	}
	c.status = finished
}

// Kill kills the process of the current command
func (c *Cmd) Kill() {
	if c.t != nil {
		c.t.Helper()
	}
	c.validateHasStarted()
	err := c.cmd.Process.Kill()
	if err != nil {
		c.fatal(err)
	}
	c.status = finished
}

// Run runs a command with name and arguments. After this, package-level
// functions will return the data about the last command run.
func Run(t *testing.T, name string, arg ...string) {
	t.Helper()
	pkgCmd = Command(t, name, arg...)
	pkgCmd.Run()
}

// Error is the command's error, if any.
func (c *Cmd) Error() error {
	if c.t != nil {
		c.t.Helper()
	}
	c.validateIsFinished()
	return c.exitError
}

// Error is the command's error, if any.
func Error() error {
	pkgCmd.t.Helper()
	return pkgCmd.Error()
}

// Stdout stream for the command
func (c *Cmd) Stdout() string {
	if c.t != nil {
		c.t.Helper()
	}
	c.validateHasStarted()
	c.stdout.mu.Lock()
	defer c.stdout.mu.Unlock()
	return c.stdout.content
}

// Stdout stream for the command
func Stdout() string {
	pkgCmd.t.Helper()
	return pkgCmd.Stdout()
}

// Stderr stream for the command
func (c *Cmd) Stderr() string {
	if c.t != nil {
		c.t.Helper()
	}
	c.validateHasStarted()
	c.stderr.mu.Lock()
	defer c.stderr.mu.Unlock()
	return c.stderr.content
}

// Stderr stream for the command
func Stderr() string {
	pkgCmd.t.Helper()
	return pkgCmd.Stderr()
}

// StdoutContains determines if command's STDOUT contains `str`, this operation
// is case insensitive.
func (c *Cmd) StdoutContains(str string) bool {
	if c.t != nil {
		c.t.Helper()
	}
	c.validateHasStarted()
	str = strings.ToLower(str)
	return retryStringTest(strings.Contains, c.stdout, str)
}

// StdoutContains determines if command's STDOUT contains `str`, this operation
// is case insensitive.
func StdoutContains(str string) bool {
	pkgCmd.t.Helper()
	return pkgCmd.StdoutContains(str)
}

// StderrContains determines if command's STDERR contains `str`, this operation
// is case insensitive.
func (c *Cmd) StderrContains(str string) bool {
	if c.t != nil {
		c.t.Helper()
	}
	c.validateHasStarted()
	str = strings.ToLower(str)
	return retryStringTest(strings.Contains, c.stderr, str)
	// return strings.Contains(strings.ToLower(c.stderr.content), str)
}

// StderrContains determines if command's STDERR contains `str`, this operation
// is case insensitive.
func StderrContains(str string) bool {
	pkgCmd.t.Helper()
	return pkgCmd.StderrContains(str)
}

// Success is a boolean status which indicates if the program exited non-zero
// or not.
func (c *Cmd) Success() bool {
	if c.t != nil {
		c.t.Helper()
	}
	c.validateIsFinished()
	return c.exitError == nil
}

// Success is a boolean status which indicates if the program exited non-zero
// or not.
func Success() bool {
	pkgCmd.t.Helper()
	return pkgCmd.Success()
}

// Failure is the inverse of Success().
func (c *Cmd) Failure() bool {
	if c.t != nil {
		c.t.Helper()
	}
	c.validateIsFinished()
	return c.exitError != nil
}

// Failure is the inverse of Success().
func Failure() bool {
	pkgCmd.t.Helper()
	return pkgCmd.Failure()
}

// StdoutMatches compares a regex to the stdout produced by the command.
func (c *Cmd) StdoutMatches(regex string) bool {
	if c.t != nil {
		c.t.Helper()
	}
	c.validateHasStarted()
	re := regexp.MustCompile(regex)
	return retryStringTest(func(got, want string) bool {
		return re.MatchString(got)
	}, c.stdout, regex)
}

// StdoutMatches compares a regex to the stdout produced by the command.
func StdoutMatches(regex string) bool {
	pkgCmd.t.Helper()
	return pkgCmd.StdoutMatches(regex)
}

// StderrMatches compares a regex to the stderr produced by the command.
func (c *Cmd) StderrMatches(regex string) bool {
	if c.t != nil {
		c.t.Helper()
	}
	c.validateHasStarted()
	re := regexp.MustCompile(regex)
	return retryStringTest(func(got, want string) bool {
		return re.MatchString(got)
	}, c.stderr, regex)
}

// StderrMatches compares a regex to the stderr produced by the command.
func StderrMatches(regex string) bool {
	pkgCmd.t.Helper()
	return pkgCmd.StderrMatches(regex)
}

// retryStringTest takes in a testFunc and will test output for the expected string until either it
// finds the expected string or times out (default 1 second)
func retryStringTest(testFunc func(string, string) bool, output *output, expected string) bool {
	ticker := time.NewTicker(100 * time.Millisecond)
	timeout := time.After(1 * time.Second)
	for {
		select {
		case <-ticker.C:
			output.mu.Lock()
			found := testFunc(strings.ToLower(output.content), expected)
			output.mu.Unlock()
			if found == true {
				return true
			}
		case <-timeout:
			return false
		}
	}
}
