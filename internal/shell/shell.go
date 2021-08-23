package shell

import (
	"bytes"
	"errors"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/mantil-io/mantil-cli/internal/log"
)

var userHome = "/tmp/home"

func Exec(args []string, dir string, successStatuses ...int) error {
	r := runner{
		dir:     dir,
		verbose: true,
		output: func(format string, v ...interface{}) {
			log.Debug(format, v...)
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
	// cmd.Env = []string{
	// 	"PATH=/usr/bin/:/bin:/usr/local/bin:/opt/bin:/var/lang/bin",
	// 	"HOME=" + userHome,
	// }

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
	//r.output("  command done exit code: %s", exitCode)
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
			//fmt.Printf("rdr.Read n = %d\n", n)
			//fmt.Printf("%s", buf[:n])
			for _, line := range strings.Split(string(buf[:n]), "\n") {
				if len(line) == 0 || line == "\n" {
					continue
				}
				//r.output("    %s", truncateTimestamp(line))
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

func AwsCli() *AwsCliExec {
	return &AwsCliExec{}
}

type AwsCliExec struct{}

func (c *AwsCliExec) SyncFrom(bucket, folder string) error {
	if err := os.MkdirAll(folder, os.ModePerm); err != nil {
		return err
	}
	return Exec([]string{"aws", "s3", "sync", bucket, folder}, "")
}

func (c *AwsCliExec) SyncTo(bucket, folder string) error {
	return Exec([]string{"aws", "s3", "sync", "--exclude", ".terraform/*", "--exclude", ".modules/*", "--exclude", ".secrets/*", "--exclude", "tfplan", "--exclude", ".DS_Store", folder, bucket}, "")
}

func FolderExists(path string) bool {
	fi, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	if err != nil {
		return false
	}
	return fi.IsDir()
}

func PrepareHome(home, secrets string) error {
	userHome = home
	awsDir := home + "/.aws"
	if FolderExists(awsDir) {
		return nil
	}
	if err := os.MkdirAll(awsDir, os.ModePerm); err != nil {
		return err
	}
	if err := Exec([]string{"cp", secrets + "/atoz/config", awsDir}, ""); err != nil {
		return err
	}
	return Exec([]string{"cp", secrets + "/atoz/credentials", awsDir}, "")
}
