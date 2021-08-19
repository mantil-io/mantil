package shell

import (
	"bytes"
	"errors"
	"io"
	"os/exec"
	"strings"

	"github.com/atoz-technology/mantil-backend/internal/log"
)

func Exec(args []string, dir string, successStatuses ...int) error {
	r := runner{
		dir:     dir,
		verbose: true,
		output: func(format string, v ...interface{}) {
			log.Info(format, v...)
		},
	}
	return r.runCmd(args)
}

type runner struct {
	verbose bool
	dir     string
	output  func(format string, v ...interface{})
}

func (r *runner) runCmd(args []string, successStatuses ...int) error {
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Dir = r.dir

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
		printCmd()
		if err := r.printToConsole(stdout); err != nil {
			return err
		}
		if err := r.printToConsole(stderr); err != nil {
			return err
		}
	}
	err = cmd.Wait()
	exitCode := exitCode(err)
	for _, ss := range successStatuses {
		if exitCode == ss {
			exitCode = 0
			err = nil
		}
	}
	if exitCode != 0 {
		if !r.verbose {
			printCmd()
		}
		r.output("FAILED with exit status %d", exitCode)
	}
	return err
}

func (r *runner) printToConsole(rdr io.ReadCloser) error {
	buf := make([]byte, 1024*16)
	for {
		n, err := rdr.Read(buf[:])
		if n > 0 {
			for _, line := range strings.Split(string(buf[:n]), "\n") {
				if len(line) == 0 || line == "\n" {
					continue
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

func Output(args []string, path string) (string, error) {
	r := runner{
		dir: path,
	}
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Dir = r.dir

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		return "", err
	}
	outStr, errStr := strings.TrimSpace(stdout.String()), strings.TrimSpace(stderr.String())
	return outStr + errStr, nil
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
