package git

import (
	"fmt"
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/mantil-io/mantil/internal/cli/log"
)

func CreateRepo(repo, path, moduleName string) error {
	_, err := git.PlainClone(path, false, &git.CloneOptions{
		URL:      repo,
		Progress: os.Stdout,
		Depth:    1,
	})
	if err != nil {
		if err == git.ErrRepositoryAlreadyExists {
			return fmt.Errorf("local repository already exists")
		}
		return err
	}
	err = os.RemoveAll(fmt.Sprintf("%s/.git", path))
	if err != nil {
		return err
	}
	if moduleName == "" {
		return nil
	}
	log.Info("Replacing module name and import paths with %s...", moduleName)
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
			return err
		}
		if info.IsDir() {
			return nil
		}
		n := info.Name()
		if strings.HasSuffix(n, ".go") || strings.HasSuffix(n, ".mod") {
			fbuf, err := ioutil.ReadFile(path)
			if err != nil {
				return err
			}
			new := strings.ReplaceAll(string(fbuf), old, new)
			err = ioutil.WriteFile(path, []byte(new), 0)
			if err != nil {
				return err
			}
		}
		return nil
	})
}
