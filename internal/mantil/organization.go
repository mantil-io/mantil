package mantil

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/atoz-technology/mantil-backend/internal/shell"
	"github.com/atoz-technology/mantil-backend/internal/template"
)

var (
	rootFolder      = "/tmp"
	templatesFolder = "/code/templates"
	modulesFolder   = "/code"
	projectsBucket  = "s3://atoz-technology-mantil-projects"
)

type Organization struct {
	Name    string
	DNSZone string
	CertArn string
}

func (o Organization) Folder() string {
	return fmt.Sprintf("%s/%s", rootFolder, o.Name)
}

func (o Organization) S3Key() string {
	return fmt.Sprintf("%s/%s", projectsBucket, o.Name)
}

func (o Organization) Pull() error {
	return shell.AwsCli().SyncFrom(o.S3Key(), o.Folder())
}

func (o Organization) ProjectFolder(name string) string {
	return fmt.Sprintf("%s/%s", o.Folder(), name)
}

func (o Organization) PushProject(name string, config interface{}) error {
	buf, err := json.MarshalIndent(config, "  ", "  ")
	if err != nil {
		return err
	}
	if err := ioutil.WriteFile(o.ProjectFolder(name)+"/config.json", buf, 0644); err != nil {
		return err
	}
	s3Key := fmt.Sprintf("%s/%s/%s", projectsBucket, o.Name, name)
	return shell.AwsCli().SyncTo(s3Key, o.ProjectFolder(name))
}

func (o Organization) renderTemplate(templateName, project string, data interface{}) error {
	return template.Exec(fmt.Sprintf("%s/%s/main.tf", templatesFolder, templateName),
		data,
		o.ProjectFolder(project)+"/main.tf",
	)
}

func (o Organization) linkModulesSecrets(project string) error {
	projectModulesFolder := o.ProjectFolder(project) + "/.modules"
	if !shell.FolderExists(projectModulesFolder) {
		if err := os.Symlink(modulesFolder, projectModulesFolder); err != nil {
			return err
		}
	}
	return nil
}

func (o Organization) PrepareProject(templateName, project string, data interface{}) error {
	if err := os.MkdirAll(o.ProjectFolder(project), os.ModePerm); err != nil {
		return err
	}
	if err := o.renderTemplate(templateName, project, data); err != nil {
		return err
	}
	return o.linkModulesSecrets(project)
}

func (o *Organization) LoadProject(project string, data interface{}) error {
	buf, err := ioutil.ReadFile(o.ProjectFolder(project) + "/config.json")
	if err != nil {
		return err
	}
	return json.Unmarshal(buf, data)
}

func (o *Organization) Load(name string) error {
	org := Organization{
		Name: name,
	}
	if err := org.Pull(); err != nil {
		return err
	}
	buf, err := ioutil.ReadFile(org.Folder() + "/config.json")
	if err != nil {
		return err
	}
	return json.Unmarshal(buf, o)
}

func (o *Organization) load(name string) error {
	org := Organization{
		Name: name,
	}
	buf, err := ioutil.ReadFile(org.Folder() + "/config.json")
	if err != nil {
		return err
	}
	return json.Unmarshal(buf, o)
}
