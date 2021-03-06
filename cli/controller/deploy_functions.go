package controller

import (
	"archive/zip"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"

	"github.com/mantil-io/mantil/cli/log"
	"github.com/mantil-io/mantil/cli/ui"
	"github.com/mantil-io/mantil/domain"
	"github.com/mantil-io/mantil/kit/shell"
)

func (d *Deploy) createMains() error {
	os.RemoveAll(filepath.Join(d.store.ProjectRoot(), FunctionsPath))
	apis, err := d.localDirs(ApiDir)
	if err != nil {
		return log.Wrap(err)
	}
	for _, api := range apis {
		dir := d.apiDir(api)
		mainDest := filepath.Join(d.apiMainDir(api), MainFile)
		if err := generateMain(api, dir, mainDest); err != nil {
			return log.Wrap(err)
		}
	}
	return nil
}

func (d *Deploy) apiDir(api string) string {
	return filepath.Join(d.store.ProjectRoot(), ApiDir, api)
}

func (d *Deploy) apiMainDir(api string) string {
	return filepath.Join(d.store.ProjectRoot(), FunctionsPath, api)
}

func (d *Deploy) localFunctions() ([]domain.Resource, error) {
	localFuncNames, err := d.localDirs(FunctionsPath)
	if err != nil {
		return nil, log.Wrap(err)
	}
	var localFuncs []domain.Resource
	for _, n := range localFuncNames {
		funcDir := d.apiMainDir(n)
		if err := d.buildTimer(func() error { return d.buildFunction(BinaryName, funcDir) }); err != nil {
			return nil, log.Wrap(err)
		}
		binaryPath := path.Join(funcDir, BinaryName)
		hash, bytes, err := fileHash(binaryPath)
		if err != nil {
			return nil, log.Wrap(err, "failed to hash %s", binaryPath)
		}
		localFuncs = append(localFuncs, domain.Resource{
			Name: n,
			Hash: hash,
		})
		log.Event(domain.Event{GoBuild: &domain.GoBuild{
			Name:     n,
			Duration: toMS(d.lastBuildDuration),
			Size:     int(bytes),
		}})
	}
	return localFuncs, nil
}

func (d *Deploy) localDirs(path string) ([]string, error) {
	files, err := ioutil.ReadDir(filepath.Join(d.store.ProjectRoot(), path))
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

func (d *Deploy) buildFunction(name, funcDir string) error {
	bl := shell.NewBufferedLogger()
	err := shell.Exec(shell.ExecOptions{
		Args:         []string{"go", "build", "-o", name, "--tags", "lambda.norpc", "--trimpath"},
		Env:          []string{"GOOS=linux", "GOARCH=arm64", "CGO_ENABLED=0"},
		WorkDir:      funcDir,
		Logger:       bl.Logger(),
		ShowExitCode: false,
		ShowShellCmd: false,
	})
	if err != nil {
		return &log.GoBuildError{Name: name, Dir: funcDir, Lines: bl.Lines()}
	}
	return nil
}

func (d *Deploy) uploadFunctions() error {
	for _, n := range d.diff.UpdatedFunctions() {
		f := d.stage.FindFunction(n)
		if f == nil {
			continue
		}
		path := filepath.Join(d.apiMainDir(n), BinaryName)
		ui.Info("\t%s", n)
		if err := d.uploadBinaryToS3(f.S3Key, path); err != nil {
			return log.Wrap(err, "failed to upload file %s to s3", path)
		}
	}
	return nil
}

func (d *Deploy) uploadBinaryToS3(key, binaryPath string) error {
	buf, err := createZipForFile(binaryPath, BinaryName)
	if err != nil {
		return err
	}
	d.uploadBytes += int64(len(buf))
	if err := d.repoPut(d.stage.Node().Bucket, key, buf); err != nil {
		return err
	}
	return nil
}

func fileHash(path string) (string, int64, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", 0, err
	}
	defer f.Close()

	h := sha256.New()
	var bytes int64
	if bytes, err = io.Copy(h, f); err != nil {
		return "", 0, err
	}
	return hex.EncodeToString(h.Sum(nil))[:HashCharacters], bytes, nil
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
