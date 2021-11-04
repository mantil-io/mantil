package controller

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/mantil-io/mantil/cli/log"
	"github.com/mantil-io/mantil/cli/ui"
	"github.com/mantil-io/mantil/domain"
	"golang.org/x/mod/sumdb/dirhash"
)

func (d *Deploy) localPublicSites() ([]domain.Resource, error) {
	localPublicNames, err := d.localDirs(PublicDir)
	if err != nil {
		return nil, log.Wrap(err)
	}
	var localPublic []domain.Resource
	for _, n := range localPublicNames {
		hash, err := d.publicSiteHash(n)
		if err != nil {
			return nil, log.Wrap(err)
		}
		localPublic = append(localPublic, domain.Resource{
			Name: n,
			Hash: hash,
		})
	}
	return localPublic, err
}

func (d *Deploy) updatePublicSiteContent() error {
	for _, u := range d.diff.UpdatedPublicSites() {
		var site *domain.PublicSite
		for _, s := range d.stage.Public.Sites {
			if s.Name == u {
				site = s
				break
			}
		}
		if site == nil {
			continue
		}
		basePath := filepath.Join(d.store.ProjectRoot(), PublicDir, site.Name)
		err := filepath.Walk(basePath, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return log.Wrap(err)
			}
			if info.IsDir() {
				return nil
			}
			relPath, err := filepath.Rel(basePath, path)
			if err != nil {
				return log.Wrap(err)
			}
			relPath = filepath.Join(site.Name, relPath)
			ui.Info("%s", relPath)
			buf, err := ioutil.ReadFile(path)
			if err != nil {
				return log.Wrap(err)
			}
			if err := d.repoPut(d.stage.Public.Bucket, relPath, buf); err != nil {
				return log.Wrap(err)
			}
			return nil
		})
		if err != nil {
			return log.Wrap(err)
		}
	}
	return nil
}

func (d *Deploy) publicSiteHash(name string) (string, error) {
	// inspired by Hash1 in dirhash but changed encoding to hex and removed prefix
	hashFunc := func(files []string, open func(string) (io.ReadCloser, error)) (string, error) {
		h := sha256.New()
		files = append([]string(nil), files...)
		sort.Strings(files)
		for _, file := range files {
			if strings.Contains(file, "\n") {
				return "", errors.New("filenames with newlines are not supported")
			}
			r, err := open(file)
			if err != nil {
				return "", err
			}
			hf := sha256.New()
			_, err = io.Copy(hf, r)
			r.Close()
			if err != nil {
				return "", err
			}
			fmt.Fprintf(h, "%x  %s\n", hf.Sum(nil), file)
		}
		return hex.EncodeToString(h.Sum(nil)), nil
	}
	dir := filepath.Join(d.store.ProjectRoot(), PublicDir, name)
	hash, err := dirhash.HashDir(dir, "", hashFunc)
	if err != nil {
		return "", log.Wrap(err)
	}
	return hash[:HashCharacters], nil
}
