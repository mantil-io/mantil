package deploy

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/mantil-io/mantil/cli/ui"
	"github.com/mantil-io/mantil/workspace"
	"golang.org/x/mod/sumdb/dirhash"
)

func (d *Cmd) publicSiteUpdates() ([]string, error) {
	var updated []string
	localSites, err := d.localDirs(PublicDir)
	if err != nil {
		return nil, err
	}
	var stageSites []string
	for _, s := range d.ctx.Stage.Public {
		stageSites = append(stageSites, s.Name)
	}
	added := diffArrays(localSites, stageSites)
	for _, a := range added {
		hash, err := d.publicSiteHash(a)
		if err != nil {
			return nil, err
		}
		d.ctx.Stage.Public = append(d.ctx.Stage.Public, &workspace.PublicSite{
			Name: a,
			Hash: hash,
		})
		updated = append(updated, a)
	}
	removed := diffArrays(stageSites, localSites)
	for _, r := range removed {
		for idx, s := range d.ctx.Stage.Public {
			if s.Name == r {
				d.ctx.Stage.Public = append(d.ctx.Stage.Public[:idx], d.ctx.Stage.Public[idx+1:]...)
			}
		}
	}
	intersection := intersectArrays(localSites, stageSites)
	for _, i := range intersection {
		hash, err := d.publicSiteHash(i)
		if err != nil {
			return nil, err
		}
		for _, s := range d.ctx.Stage.Public {
			if s.Name == i && hash != s.Hash {
				s.Hash = hash
				updated = append(updated, i)
			}
		}
	}
	return updated, nil
}

func (d *Cmd) updatePublicSiteContent() error {
	for _, u := range d.updatedPublicSites {
		var site *workspace.PublicSite
		for _, s := range d.ctx.Stage.Public {
			if s.Name == u {
				site = s
				break
			}
		}
		if site == nil {
			continue
		}
		ui.Info("updating public site %s", site.Name)
		basePath := filepath.Join(d.ctx.Path, PublicDir, site.Name)
		err := filepath.Walk(basePath, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				return nil
			}
			relPath, err := filepath.Rel(basePath, path)
			if err != nil {
				return err
			}
			ui.Info("uploading file %s...", relPath)
			buf, err := ioutil.ReadFile(path)
			if err != nil {
				return err
			}
			if err := d.awsClient.PutObjectToS3Bucket(site.Bucket, relPath, buf); err != nil {
				return err
			}
			return nil
		})
		if err != nil {
			return err
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
