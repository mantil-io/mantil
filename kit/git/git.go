package git

import (
	"fmt"
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/pkg/errors"
)

var ErrRepositoryNotFound = fmt.Errorf("repository not found")

func CreateRepo(repo, path, moduleName string) error {
	co := &git.CloneOptions{
		URL:   repo,
		Depth: 1,
	}
	_, err := git.PlainClone(path, false, co)
	if err != nil {
		if err == transport.ErrRepositoryNotFound {
			return errors.WithStack(ErrRepositoryNotFound)
		}
		return errors.WithStack(err)
	}
	err = os.RemoveAll(fmt.Sprintf("%s/.git", path))
	if err != nil {
		return errors.WithStack(err)
	}
	if moduleName == "" {
		return nil
	}
	return replaceImportPaths(path, repo, moduleName)
}

func DeleteRepo(path string) error {
	return os.RemoveAll(path)
}

func replaceImportPaths(repoDir, old, new string) error {
	old = strings.ReplaceAll(old, "https://", "")
	new = strings.ReplaceAll(new, "https://", "")
	return filepath.Walk(repoDir, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return errors.WithStack(err)
		}
		if info.IsDir() {
			return nil
		}
		n := info.Name()
		if strings.HasSuffix(n, ".go") || strings.HasSuffix(n, ".mod") {
			fbuf, err := ioutil.ReadFile(path)
			if err != nil {
				return errors.WithStack(err)
			}
			new := strings.ReplaceAll(string(fbuf), old, new)
			err = ioutil.WriteFile(path, []byte(new), 0)
			if err != nil {
				return errors.WithStack(err)
			}
		}
		return nil
	})
}
