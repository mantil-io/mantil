package deploy

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/mantil-io/mantil/internal/cli/log"
	"github.com/mantil-io/mantil/internal/mantil"
	"github.com/mantil-io/mantil/internal/util"
	"golang.org/x/mod/sumdb/dirhash"
)

func (d *DeployCmd) staticSiteUpdates() ([]mantil.ProjectUpdate, error) {
	var updates []mantil.ProjectUpdate
	localSites, err := d.localDirs(StaticSitesDir)
	if err != nil {
		return nil, err
	}
	var projectSites []string
	for _, s := range d.project.StaticWebsites {
		projectSites = append(projectSites, s.Name)
	}
	added := util.DiffArrays(localSites, projectSites)
	for _, a := range added {
		hash, err := d.staticWebsiteHash(a)
		if err != nil {
			return nil, err
		}
		updates = append(updates, mantil.ProjectUpdate{
			StaticWebsite: &mantil.StaticWebsiteUpdate{
				Name: a,
				Hash: hash,
			},
			Action: mantil.Add,
		})
	}
	removed := util.DiffArrays(projectSites, localSites)
	for _, r := range removed {
		updates = append(updates, mantil.ProjectUpdate{
			StaticWebsite: &mantil.StaticWebsiteUpdate{
				Name: r,
			},
			Action: mantil.Remove,
		})
	}
	intersection := util.IntersectArrays(localSites, projectSites)
	for _, i := range intersection {
		hash, err := d.staticWebsiteHash(i)
		if err != nil {
			return nil, err
		}
		for _, s := range d.project.StaticWebsites {
			if s.Name == i && hash != s.Hash {
				updates = append(updates, mantil.ProjectUpdate{
					StaticWebsite: &mantil.StaticWebsiteUpdate{
						Name: i,
						Hash: hash,
					},
					Action: mantil.Update,
				})
			}
		}
	}
	return updates, nil
}

func (d *DeployCmd) updateStaticWebsiteContents() error {
	for _, u := range d.updates {
		if u.StaticWebsite == nil || (u.Action != mantil.Add && u.Action != mantil.Update) {
			continue
		}
		var site *mantil.StaticWebsite
		for _, s := range d.project.StaticWebsites {
			if s.Name == u.StaticWebsite.Name {
				site = &s
				break
			}
		}
		if site == nil {
			continue
		}
		log.Info("updating static website %s", site.Name)
		basePath := filepath.Join(d.path, StaticSitesDir, site.Name)
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
			log.Info("uploading file %s...", relPath)
			buf, err := ioutil.ReadFile(path)
			if err != nil {
				return err
			}
			if err := d.aws.PutObjectToS3Bucket(site.Bucket, relPath, buf); err != nil {
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

func (d *DeployCmd) staticWebsiteHash(name string) (string, error) {
	dir := filepath.Join(d.path, StaticSitesDir, name)
	hash, err := dirhash.HashDir(dir, "", dirhash.Hash1)
	if err != nil {
		return "", err
	}
	return hash, nil
}
