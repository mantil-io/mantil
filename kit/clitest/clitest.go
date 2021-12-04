// fetures:
// * snima stdout i stderr na disk kako se koji pojavi
// * vraca nekog tko radi asserts rezultat: stdout, stderr, exitcode \
// * moze se postaviti
//   * working folder
//   * env
package clitest

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"testing"
	"time"
)

var (
	showOnce          sync.Once
	outputDir         string
	commandOutputPath func(testName, cmdStr string) string
)

func Show() {
	files, err := outputFiles(outputDir)
	if err != nil {
		fmt.Printf("Error: %s", err)
		return
	}
	for _, file := range files {
		buf, _ := ioutil.ReadFile(file)
		fmt.Printf("%s\n%s\n", file, buf)
	}
}

func outputFiles(root string) ([]string, error) {
	var files []string
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			files = append(files, path)
		}
		return nil
	})
	return files, err
}

func init() {
	var err error
	outputDir, err = ioutil.TempDir("", "clitest-output-*")
	if err != nil {
		panic(err)
	}

	var cnt int
	var mu sync.Mutex

	commandOutputPath = func(testName, cmdStr string) string {
		mu.Lock()
		cnt++
		c := cnt
		mu.Unlock()

		fn := fmt.Sprintf("%s-%03d-%s", testName, c, strings.Replace(cmdStr, " ", "_", -1))

		// replace all non-alphanumeric characters
		reg, err := regexp.Compile("[^A-Za-z0-9-_]+")
		if err != nil {
			panic(err)
		}
		fn = reg.ReplaceAllString(fn, "_")
		return filepath.Join(outputDir, fn)
	}
}

type TestingT interface {
	Helper()
	Errorf(format string, args ...interface{})
	Logf(format string, args ...interface{})
	Deadline() (deadline time.Time, ok bool)
	Name() string
}

type Env struct {
	t              TestingT // using instead of so can be use without it *testing.T
	workdir        string
	vars           []string
	cmdStr         string
	stdoutFilename string
	stderrFilename string
}

func New(t TestingT) *Env {
	t.Helper()

	showOnce.Do(func() {
		if testing.Verbose() {
			fmt.Printf("cli output dir %s\n", outputDir)
		}
	})

	return &Env{
		t:    t,
		vars: os.Environ(),
	}
}

func (e *Env) Env(k, v string) *Env {
	n, _ := e.varsWithout(k)
	e.vars = n
	e.vars = append(e.vars, fmt.Sprintf("%s=%s", k, v))
	return e
}

func (e *Env) Path(v string) *Env {
	n, p := e.varsWithout("PATH")
	p = v + ":" + p
	e.vars = n
	e.vars = append(e.vars, fmt.Sprintf("PATH=%s", p))
	return e
}

func (e *Env) varsWithout(key string) ([]string, string) {
	var n []string
	var val string
	key = key + "="
	for _, e := range e.vars {
		if strings.HasPrefix(e, key) {
			val = strings.Split(e, "=")[1]
			continue
		}
		n = append(n, e)
	}
	return n, val
}

func (e *Env) Workdir(wd ...string) *Env {
	if len(wd) == 1 {
		e.workdir = wd[0]
		return e
	}
	e.workdir = filepath.Join(wd...)
	return e
}

func (e *Env) GetWorkdir() string {
	return e.workdir
}

func (e *Env) Cd(folderName string) string {
	e.workdir = filepath.Join(e.workdir, folderName)
	return e.workdir
}

func (e *Env) CpToWorkdir(from, to string) {
	err := Cp(from, filepath.Join(e.workdir, to))
	if err != nil {
		e.t.Errorf("failed to copy %s to %s error: %s", from, to, err)
	}

}

func (e *Env) Run(name string, arg ...string) *Expect {
	e.t.Helper()
	// prepare
	e.cmdStr = fmt.Sprintf("$%s %s", name, strings.Join(arg, " "))
	outputPath := commandOutputPath(e.t.Name(), e.cmdStr)
	e.stdoutFilename = outputPath + ".out"
	e.stderrFilename = outputPath + ".err"
	start := time.Now()
	// atual command
	cmd := exec.Command(name, arg...)

	// show start of the command only if some time is running (we have end log with duration)
	done := make(chan struct{})
	go func() {
		select {
		case <-time.After(time.Second):
			e.logf("%s", e.cmdStr)
		case <-done:
			return
		}
	}()

	// kill running process when test timeouts
	go func() {
		deadline, ok := e.t.Deadline()
		if !ok {
			//it is ok if test flag is set and if not
			deadline = time.Now().Add(time.Minute)
		}
		ctx, cancel := context.WithDeadline(context.Background(), deadline.Add(-time.Second))
		defer cancel()
		select {
		case <-ctx.Done():
			err := cmd.Process.Kill()
			if err == nil {
				e.logf("[%s] killed", e.cmdStr)
			}
			return
		case <-done:
			return
		}
	}()

	if e.workdir != "" {
		cmd.Dir = e.workdir
	}
	if e.vars != nil {
		cmd.Env = e.vars
	}

	// redirect stdout and stderr to files
	out, err := os.Create(e.stdoutFilename)
	if err == nil {
		cmd.Stdout = out
	} else {
		e.t.Logf("can't create file %s", e.stdoutFilename)
	}
	eut, err := os.Create(e.stderrFilename)
	if err == nil {
		cmd.Stderr = eut
	} else {
		e.t.Logf("can't create file %s", e.stderrFilename)
	}

	// RUN
	runError := cmd.Run()

	// close output files, signal done, log time
	close(done)
	out.Close()
	eut.Close()
	dur := time.Now().Sub(start).Round(time.Millisecond)
	e.logf("%s done in %v", e.cmdStr, dur)

	ex := &Expect{
		t:        e.t,
		cmdStr:   e.cmdStr,
		cmd:      cmd,
		runError: runError,
		stdout:   fileContent(e.stdoutFilename),
		stderr:   fileContent(e.stderrFilename),
		exitCode: exitCode(runError),
	}
	ex.Stdout()
	return ex
}

func fileContent(fn string) string {
	buf, err := ioutil.ReadFile(fn)
	if err == nil {
		if len(buf) == 0 {
			os.Remove(fn)
		}
		return string(buf)
	}
	return ""
}

func exitCode(err error) int {
	if err == nil {
		return 0
	}
	var ee *exec.ExitError
	if errors.As(err, &ee) {
		return ee.ExitCode()
	}
	return math.MinInt
}

func (e *Env) logf(format string, arg ...interface{}) {
	if !testing.Verbose() {
		return
	}
	e.t.Helper()
	format = fmt.Sprintf("    %-20s ", e.t.Name()) + format + "\n"
	fmt.Printf(format, arg...)
}

type Expect struct {
	t        TestingT
	cmdStr   string
	cmd      *exec.Cmd
	runError error
	stdout   string
	stderr   string
	exitCode int
	out      string
	outType  string
}

func (e *Expect) Success() *Expect {
	e.t.Helper()
	if e.exitCode != 0 {
		if e.exitCode == math.MinInt {
			e.t.Errorf("[%s] error %s", e.cmdStr, e.runError)
		} else {
			e.t.Errorf("[%s] should have zero exit code, acutal %d", e.cmdStr, e.exitCode)
		}
	}
	return e
}

func (e *Expect) Fail() *Expect {
	e.t.Helper()
	if e.exitCode == 0 {
		e.t.Errorf("[%s] should have non zero exit code", e.cmdStr)
	}
	return e
}

func (e *Expect) Stdout() *Expect {
	e.t.Helper()
	e.out = e.stdout
	e.outType = "stdout"
	return e
}

func (e *Expect) Stderr() *Expect {
	e.t.Helper()
	e.out = e.stderr
	e.outType = "stderr"
	return e
}

func (e *Expect) Contains(str string) *Expect {
	e.t.Helper()
	if !strings.Contains(e.out, str) {
		e.t.Errorf("[%s] %s should contain %s", e.cmdStr, e.outType, str)
	}
	return e
}

func (e *Expect) GetStdout() string {
	e.t.Helper()
	return e.stdout
}

// Copy the src file to dst. Any existing file will be overwritten and will not
// copy file attributes.
func Cp(src, dst string) error {
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

var globalMu sync.Mutex

func (e *Env) WithWorkdir(cb func()) {
	globalMu.Lock()
	defer globalMu.Unlock()
	cwd, err := os.Getwd()
	os.Chdir(e.workdir)
	cb()
	if err == nil {
		os.Chdir(cwd)
	}
}
