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

	"github.com/mantil-io/mantil/cli/ui"
	"github.com/mantil-io/mantil/shell"
	"github.com/mantil-io/mantil/workspace"
)

func (d *Cmd) functionUpdates() (resourceDiff, error) {
	var diff resourceDiff
	localFuncs, err := d.localDirs(FunctionsDir)
	if err != nil {
		return diff, err
	}
	var stageFuncs []string
	for _, f := range d.ctx.Stage.Functions {
		stageFuncs = append(stageFuncs, f.Name)
	}
	diff.added = diffArrays(localFuncs, stageFuncs)
	for _, a := range diff.added {
		if !workspace.FunctionNameAvailable(a) {
			return diff, fmt.Errorf("api name \"%s\" is reserved", a)
		}
		d.ctx.Stage.Functions = append(d.ctx.Stage.Functions, &workspace.Function{
			Name: a,
		})
	}
	diff.removed = diffArrays(stageFuncs, localFuncs)
	for _, r := range diff.removed {
		for idx, sf := range d.ctx.Stage.Functions {
			if sf.Name == r {
				d.ctx.Stage.Functions = append(d.ctx.Stage.Functions[:idx], d.ctx.Stage.Functions[idx+1:]...)
				break
			}
		}
	}
	diff.updated = d.prepareFunctionsForDeploy()
	return diff, nil
}

func (d *Cmd) localDirs(path string) ([]string, error) {
	files, err := ioutil.ReadDir(filepath.Join(d.ctx.Path, path))
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
func (d *Cmd) prepareFunctionsForDeploy() []string {
	var updatedFunctions []string
	for _, f := range d.ctx.Stage.Functions {
		ui.Info("building function %s", f.Name)
		funcDir := path.Join(d.ctx.Path, FunctionsDir, f.Name)
		if err := d.buildFunction(BinaryName, funcDir); err != nil {
			ui.Errorf("skipping function %s due to error while building - %v", f.Name, err)
			continue
		}
		binaryPath := path.Join(funcDir, BinaryName)
		hash, err := fileHash(binaryPath)
		if err != nil {
			ui.Errorf("skipping function %s due to error while calculating binary hash - %v", f.Name, err)
			continue
		}
		if hash != f.Hash {
			updatedFunctions = append(updatedFunctions, f.Name)
			f.Hash = hash
			ui.Debug("creating function %s as zip package type", f.Name)
			f.SetS3Key(fmt.Sprintf("%s/functions/%s-%s.zip", workspace.StageBucketPrefix(d.ctx.Project.Name, d.ctx.Stage.Name), f.Name, f.Hash))
			ui.Debug("uploading function %s to s3", f.Name)
			if err := d.uploadBinaryToS3(f.S3Key, binaryPath); err != nil {
				ui.Errorf("skipping function %s due to error while processing s3 file - %v", f.Name, err)
				continue
			}
		}
	}
	return updatedFunctions
}

func (d *Cmd) buildFunction(name, funcDir string) error {
	return shell.Exec(shell.ExecOptions{
		Args:    []string{"env", "GOOS=linux", "GOARCH=amd64", "go", "build", "-o", name, "--tags", "lambda.norpc"},
		WorkDir: funcDir,
		Logger:  ui.Debug,
	})
}

func (d *Cmd) uploadBinaryToS3(key, binaryPath string) error {
	buf, err := createZipForFile(binaryPath, BinaryName)
	if err != nil {
		return err
	}
	if err := d.awsClient.PutObjectToS3Bucket(d.ctx.Account.Bucket, key, buf); err != nil {
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
