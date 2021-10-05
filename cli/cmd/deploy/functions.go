package deploy

import (
	"archive/zip"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path"

	"github.com/mantil-io/mantil/cli/ui"
	"github.com/mantil-io/mantil/shell"
	"github.com/mantil-io/mantil/workspace"
)

func (d *DeployCmd) functionUpdates() (updated bool, err error) {
	localFuncs, err := d.localDirs(FunctionsDir)
	if err != nil {
		return false, err
	}
	var stageFuncs []string
	for _, f := range d.ctx.Stage.Functions {
		stageFuncs = append(stageFuncs, f.Name)
	}
	added := diffArrays(localFuncs, stageFuncs)
	for _, a := range added {
		if !workspace.FunctionNameAvailable(a) {
			return false, fmt.Errorf("api name \"%s\" is reserved", a)
		}
		d.ctx.Stage.Functions = append(d.ctx.Stage.Functions, &workspace.Function{
			Name: a,
		})
	}
	removed := diffArrays(stageFuncs, localFuncs)
	for _, r := range removed {
		for idx, sf := range d.ctx.Stage.Functions {
			if sf.Name == r {
				d.ctx.Stage.Functions = append(d.ctx.Stage.Functions[:idx], d.ctx.Stage.Functions[idx+1:]...)
				break
			}
		}
	}
	updated = d.prepareFunctionsForDeploy()
	updated = updated || len(added) > 0 || len(removed) > 0
	return updated, nil
}

// prepareFunctionsForDeploy goes through stage functions, checks which ones have changed
// and uploads new version to s3 if necessary
func (d *DeployCmd) prepareFunctionsForDeploy() (updated bool) {
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
			updated = true
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
	return updated
}

func (d *DeployCmd) buildFunction(name, funcDir string) error {
	return shell.Exec(shell.ExecOptions{
		Args:    []string{"env", "GOOS=linux", "GOARCH=amd64", "go", "build", "-o", name, "--tags", "lambda.norpc"},
		WorkDir: funcDir,
		Logger:  ui.Debug,
	})
}

func (d *DeployCmd) uploadBinaryToS3(key, binaryPath string) error {
	buf, err := createZipForFile(binaryPath, BinaryName)
	if err != nil {
		return err
	}
	if err := d.aws.PutObjectToS3Bucket(d.ctx.Account.Bucket, key, buf); err != nil {
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
