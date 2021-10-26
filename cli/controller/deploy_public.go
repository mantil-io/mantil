package controller

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/mantil-io/mantil/cli/log"
	"github.com/mantil-io/mantil/cli/ui"
	"github.com/mantil-io/mantil/workspace"
	"golang.org/x/mod/sumdb/dirhash"
)

func (d *Deploy) localPublicSites() ([]workspace.Resource, error) {
	localPublicNames, err := d.localDirs(PublicDir)
	if err != nil {
		return nil, log.Wrap(err)
	}
	var localPublic []workspace.Resource
	for _, n := range localPublicNames {
		hash, err := d.publicSiteHash(n)
		if err != nil {
			return nil, log.Wrap(err)
		}
		localPublic = append(localPublic, workspace.Resource{
			Name: n,
			Hash: hash,
		})
	}
	return localPublic, err
}

func (d *Deploy) updatePublicSiteContent() error {
	for _, u := range d.diff.UpdatedPublicSites() {
		var site *workspace.PublicSite
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
			if err := d.awsClient.PutObjectToS3Bucket(d.stage.Public.Bucket, relPath, buf); err != nil {
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
	dir := filepath.Join(d.store.ProjectRoot(), PublicDir, name)
	hash, err := dirhash.HashDir(dir, "", dirhash.Hash1)
	if err != nil {
		return "", err
	}
	return hash, nil
}
