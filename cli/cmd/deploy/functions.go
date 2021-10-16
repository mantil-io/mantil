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
)

func (d *Cmd) functionUpdates() (resourceDiff, error) {
	var diff resourceDiff
	localFuncs, err := d.localDirs(FunctionsDir)
	if err != nil {
		return diff, err
	}
	var stageFuncs []string
	for _, f := range d.stage.Functions {
		stageFuncs = append(stageFuncs, f.Name)
	}
	diff.added = diffArrays(localFuncs, stageFuncs)
	if rerr := d.stage.AddFunctions(diff.added); rerr != nil {
		return diff, log.WithUserMessage(rerr, "\"%s\" is reserved name", rerr.Name)
	}
	diff.removed = diffArrays(stageFuncs, localFuncs)
	d.stage.RemoveFunctions(diff.removed)
	diff.updated, err = d.prepareFunctionsForDeploy()
	return diff, err
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

// prepareFunctionsForDeploy goes through stage functions, checks which ones have changed
// and uploads new version to s3 if necessary
func (d *Cmd) prepareFunctionsForDeploy() ([]string, error) {
	var updatedFunctions []string
	d.functionsForUpload = make([]uploadData, 0)
	for _, f := range d.stage.Functions {
		ui.Info("%s", f.Name)
		funcDir := path.Join(d.path, FunctionsDir, f.Name)
		if err := d.buildTimer(func() error { return d.buildFunction(BinaryName, funcDir) }); err != nil {
			return nil, log.Wrap(err)
		}

		binaryPath := path.Join(funcDir, BinaryName)
		hash, err := fileHash(binaryPath)
		if err != nil {
			return nil, log.WithUserMessage(err, fmt.Sprintf("Hashing %s failed", binaryPath))
		}
		if hash != f.Hash {
			updatedFunctions = append(updatedFunctions, f.Name)
			f.SetHash(hash)
			d.functionsForUpload = append(d.functionsForUpload, uploadData{
				name:       f.Name,
				s3Key:      f.S3Key,
				binaryPath: binaryPath,
			})
		}
	}
	return updatedFunctions, nil
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

func (d *Cmd) upload() error {
	for _, f := range d.functionsForUpload {
		ui.Info(f.name)
		if err := d.uploadBinaryToS3(f.s3Key, f.binaryPath); err != nil {
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

// returns a1 - a2
func diffArrays(a1 []string, a2 []string) []string {
	m := make(map[string]bool)
	for _, e := range a2 {
		m[e] = true
	}
	var diff []string
	for _, e := range a1 {
		if m[e] {
			continue
		}
		diff = append(diff, e)
	}
	return diff
}

// returns a1 n a2
func intersectArrays(a1 []string, a2 []string) []string {
	m := make(map[string]bool)
	for _, e := range a1 {
		m[e] = true
	}
	var intersection []string
	for _, e := range a2 {
		if m[e] {
			intersection = append(intersection, e)
		}
	}
	return intersection
}
