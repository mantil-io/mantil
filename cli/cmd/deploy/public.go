package deploy

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/mantil-io/mantil/cli/log"
	"github.com/mantil-io/mantil/cli/ui"
	"github.com/mantil-io/mantil/workspace"
	"golang.org/x/mod/sumdb/dirhash"
)

func (d *Cmd) publicSiteUpdates() (resourceDiff, error) {
	var diff resourceDiff
	localSites, err := d.localDirs(PublicDir)
	if err != nil {
		return diff, err
	}
	var stageSites []string
	for _, s := range d.ctx.Stage.Public.Sites {
		stageSites = append(stageSites, s.Name)
	}
	diff.added = diffArrays(localSites, stageSites)
	for _, a := range diff.added {
		hash, err := d.publicSiteHash(a)
		if err != nil {
			return diff, err
		}
		d.ctx.Stage.Public.Sites = append(d.ctx.Stage.Public.Sites, &workspace.PublicSite{
			Name: a,
			Hash: hash,
		})
		diff.updated = append(diff.updated, a)
	}
	diff.removed = diffArrays(stageSites, localSites)
	for _, r := range diff.removed {
		for idx, s := range d.ctx.Stage.Public.Sites {
			if s.Name == r {
				d.ctx.Stage.Public.Sites = append(d.ctx.Stage.Public.Sites[:idx], d.ctx.Stage.Public.Sites[idx+1:]...)
			}
		}
	}
	intersection := intersectArrays(localSites, stageSites)
	for _, i := range intersection {
		hash, err := d.publicSiteHash(i)
		if err != nil {
			return diff, err
		}
		for _, s := range d.ctx.Stage.Public.Sites {
			if s.Name == i && hash != s.Hash {
				s.Hash = hash
				diff.updated = append(diff.updated, i)
			}
		}
	}
	return diff, nil
}

func (d *Cmd) updatePublicSiteContent() error {
	for _, u := range d.publicDiff.updated {
		var site *workspace.PublicSite
		for _, s := range d.ctx.Stage.Public.Sites {
			if s.Name == u {
				site = s
				break
			}
		}
		if site == nil {
			continue
		}
		basePath := filepath.Join(d.ctx.Path, PublicDir, site.Name)
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
			if err := d.awsClient.PutObjectToS3Bucket(d.ctx.Stage.Public.Bucket, relPath, buf); err != nil {
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

func (d *Cmd) publicSiteHash(name string) (string, error) {
	dir := filepath.Join(d.ctx.Path, PublicDir, name)
	hash, err := dirhash.HashDir(dir, "", dirhash.Hash1)
	if err != nil {
		return "", err
	}
	return hash, nil
}
