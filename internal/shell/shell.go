package shell

import (
	"errors"
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
}

func Exec(opt ExecOptions) error {
	r := runner{
		verbose: true,
		output:  opt.Logger,
		opt:     opt,
	}
	if opt.Logger == nil {
		var std = log.New(os.Stderr, log.Prefix(), 0)
		r.output = func(format string, v ...interface{}) {
			std.Printf(format, v...)
		}
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
	verbose bool
	env     []string
	output  func(format string, v ...interface{})
	err     error
	opt     ExecOptions
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
		r.output(">> %s", strings.Join(args, " "))
	}
	if err := cmd.Start(); err != nil {
		return err
	}
	if r.verbose {
		if r.opt.ShowShellCmd {
			printCmd()
		}
		if err := r.catchOutput(stdout); err != nil {
			return err
		}
		if err := r.catchOutput(stderr); err != nil {
			return err
		}
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
	if exitCode != 0 {
		if !r.opt.ShowShellCmd {
			printCmd()
		}
		r.output("FAILED with exit status %d", exitCode)
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
				r.output("%s", line)
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
