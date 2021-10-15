package shell

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"
)

type ExecOptions struct {
	Args           []string
	Env            []string
	WorkDir        string
	SucessStatuses []int
	Logger         func(format string, v ...interface{})
	ErrorsMap      map[string]error
	ShowShellCmd   bool
	ShowExitCode   bool
}

func StdLogger() func(format string, v ...interface{}) {
	var std = log.New(os.Stderr, log.Prefix(), 0)
	return func(format string, v ...interface{}) {
		std.Printf(format, v...)
	}
}

type BufferedLogger struct {
	lines []string
}

func NewBufferedLogger() *BufferedLogger {
	return &BufferedLogger{}
}

func (b *BufferedLogger) Logger() func(format string, v ...interface{}) {
	return func(format string, v ...interface{}) {
		b.lines = append(b.lines, fmt.Sprintf(format, v...))
	}
}

func (b *BufferedLogger) Lines() []string {
	return b.lines
}

func Exec(opt ExecOptions) error {
	r := runner{
		logger: opt.Logger,
		opt:    opt,
	}
	if opt.Logger == nil {
		r.logger = StdLogger()
	}
	r.env = append(os.Environ(), opt.Env...)
	return r.runCmd(opt.Args, opt.SucessStatuses...)
}

func addCurrentPath(env []string) []string {
	for _, s := range env {
		if strings.HasPrefix(s, "PATH") {
			return env
		}
	}
	val, ok := os.LookupEnv("PATH")
	if !ok {
		return env
	}
	return append(env, "PATH="+val)
}

type runner struct {
	env    []string
	logger func(format string, v ...interface{})
	err    error
	opt    ExecOptions
}

func (r *runner) runCmd(args []string, successStatuses ...int) error {
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Dir = r.opt.WorkDir
	if r.env != nil {
		cmd.Env = r.env
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}

	printCmd := func() {
		r.logger(">> %s", strings.Join(args, " "))
	}
	if err := cmd.Start(); err != nil {
		return err
	}

	if r.opt.ShowShellCmd {
		printCmd()
	}
	if err := r.catchOutput(stdout); err != nil {
		return err
	}
	if err := r.catchOutput(stderr); err != nil {
		return err
	}

	err = cmd.Wait()
	exitCode := exitCode(err)
	//r.output("  command done exit code: %s", exitCode)
	for _, ss := range successStatuses {
		if exitCode == ss {
			exitCode = 0
			err = nil
		}
	}
	if exitCode != 0 && r.opt.ShowExitCode {
		if !r.opt.ShowShellCmd {
			printCmd()
		}
		r.logger("FAILED with exit code %d", exitCode)
	}
	if r.err != nil {
		return r.err
	}
	return err
}

func (r *runner) catchOutput(rdr io.ReadCloser) error {
	buf := make([]byte, 1024*16)
	for {
		n, err := rdr.Read(buf[:])
		if n > 0 {
			//fmt.Printf("rdr.Read n = %d\n", n)
			//fmt.Printf("%s", buf[:n])
			for _, line := range strings.Split(string(buf[:n]), "\n") {
				if len(line) == 0 || line == "\n" {
					continue
				}
				// find error in line
				if r.err == nil && len(r.opt.ErrorsMap) > 0 {
					for k, v := range r.opt.ErrorsMap {
						if strings.Contains(line, k) {
							r.err = v
						}
					}
				}
				r.logger("%s", line)
			}
		}
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
	}
}

func exitCode(err error) int {
	if err == nil {
		return 0
	}
	var ee *exec.ExitError
	if errors.As(err, &ee) {
		return ee.ExitCode()
	}
	return 127
}

// TODO: potentially incorporate this into package above
// quick fix just so the functions can build for now
func Output(opt ExecOptions) (string, error) {
	args := opt.Args
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Dir = opt.WorkDir
	if opt.Env != nil {
		cmd.Env = append(os.Environ(), opt.Env...)
	}

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	outStr, errStr := strings.TrimSpace(stdout.String()), strings.TrimSpace(stderr.String())
	return outStr + errStr, err
}
