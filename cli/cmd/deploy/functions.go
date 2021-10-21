package deploy

import (
	"archive/zip"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/mantil-io/mantil/cli/log"
	"github.com/mantil-io/mantil/cli/ui"
	"github.com/mantil-io/mantil/shell"
	"github.com/mantil-io/mantil/workspace"
)

func (d *Cmd) localFunctions() ([]workspace.Resource, error) {
	localFuncNames, err := d.localDirs(FunctionsDir)
	if err != nil {
		return nil, log.Wrap(err)
	}
	var localFuncs []workspace.Resource
	for _, n := range localFuncNames {
		ui.Info(n)
		funcDir := path.Join(d.path, FunctionsDir, n)
		if err := d.buildTimer(func() error { return d.buildFunction(BinaryName, funcDir) }); err != nil {
			return nil, log.Wrap(err)
		}
		binaryPath := path.Join(funcDir, BinaryName)
		hash, err := fileHash(binaryPath)
		if err != nil {
			return nil, log.WithUserMessage(err, fmt.Sprintf("Hashing %s failed", binaryPath))
		}
		localFuncs = append(localFuncs, workspace.Resource{
			Name: n,
			Hash: hash,
		})
	}
	return localFuncs, nil
}

func (d *Cmd) localDirs(path string) ([]string, error) {
	files, err := ioutil.ReadDir(filepath.Join(d.path, path))
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	dirs := []string{}
	for _, f := range files {
		if !f.IsDir() {
			continue
		}
		dirs = append(dirs, f.Name())
	}
	return dirs, nil
}

func (d *Cmd) buildFunction(name, funcDir string) error {
	bl := shell.NewBufferedLogger()
	err := shell.Exec(shell.ExecOptions{
		Args:         []string{"env", "GOOS=linux", "GOARCH=amd64", "go", "build", "-o", name, "--tags", "lambda.norpc"},
		WorkDir:      funcDir,
		Logger:       bl.Logger(),
		ShowExitCode: false,
		ShowShellCmd: false,
	})
	if err != nil {
		return log.WithUserMessage(err, strings.Join(bl.Lines(), "\n"))
	}
	return err
}

func (d *Cmd) uploadFunctions() error {
	for _, n := range d.diff.UpdatedFunctions() {
		f := d.stage.FindFunction(n)
		if f == nil {
			continue
		}
		path := filepath.Join(d.path, FunctionsDir, n, BinaryName)
		ui.Info(n)
		if err := d.uploadBinaryToS3(f.S3Key, path); err != nil {
			return log.WithUserMessage(err, "Failed to upload file to s3")
		}
	}
	return nil
}

func (d *Cmd) uploadBinaryToS3(key, binaryPath string) error {
	buf, err := createZipForFile(binaryPath, BinaryName)
	if err != nil {
		return err
	}
	d.uploadBytes += int64(len(buf))
	if err := d.awsClient.PutObjectToS3Bucket(d.stage.Account().Bucket, key, buf); err != nil {
		return err
	}
	return nil
}

func fileHash(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

func createZipForFile(path, name string) ([]byte, error) {
	buf := new(bytes.Buffer)
	w := zip.NewWriter(buf)

	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return nil, err
	}

	hdr, err := zip.FileInfoHeader(info)
	if err != nil {
		return nil, err
	}

	// using base name in the header so zip doesn't create a directory
	hdr.Name = name
	hdr.Method = zip.Deflate
	dst, err := w.CreateHeader(hdr)
	if err != nil {
		return nil, err
	}

	_, err = io.Copy(dst, file)
	if err != nil {
		return nil, err
	}

	if err := w.Close(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
