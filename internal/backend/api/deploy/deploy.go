package deploy

import (
	"encoding/json"
	"fmt"

	"github.com/mantil-io/mantil/internal/aws"
	"github.com/mantil-io/mantil/internal/backend/assets"
	"github.com/mantil-io/mantil/internal/backend/log"
	"github.com/mantil-io/mantil/internal/backend/terraform"
	"github.com/mantil-io/mantil/internal/mantil"
)

const (
	FunctionsDir = "functions"
	BuildDir     = "build"
)

type Deploy struct {
	aws     *aws.AWS
	project *mantil.Project
	updates []mantil.ProjectUpdate
	tf      *terraform.Terraform
}

func New(project *mantil.Project, updates []mantil.ProjectUpdate, tf *terraform.Terraform) (*Deploy, error) {
	awsClient, err := aws.New()
	if err != nil {
		return nil, err
	}
	assets.StartServer()
	return &Deploy{
		aws:     awsClient,
		project: project,
		updates: updates,
		tf:      tf,
	}, nil
}

func (d *Deploy) updateProject() {
	for _, u := range d.updates {
		switch {
		case u.Function != nil:
			d.processFunctionUpdate(u.Function, u.Action)
		case u.StaticWebsite != nil:
			d.processStaticWebsiteUpdate(u.StaticWebsite, u.Action)
		}
	}
	d.project.AddFunctionDefaults()
}

func (d *Deploy) processFunctionUpdate(fu *mantil.FunctionUpdate, action mantil.UpdateAction) {
	var idx int = -1
	for i, f := range d.project.Functions {
		if fu.Name == f.Name {
			idx = i
			break
		}
	}
	switch action {
	case mantil.Update:
		if idx == -1 {
			return
		}
		f := d.project.Functions[idx]
		f.Hash = fu.Hash
		f.S3Key = fu.S3Key
		d.project.Functions[idx] = f
	case mantil.Add:
		f := mantil.Function{
			Name:  fu.Name,
			Hash:  fu.Hash,
			S3Key: fu.S3Key,
		}
		d.project.Functions = append(d.project.Functions, f)
	case mantil.Remove:
		if idx == -1 {
			return
		}
		fs := d.project.Functions
		fs = append(fs[:idx], fs[idx+1:]...)
		d.project.Functions = fs
	}
}

func (d *Deploy) processStaticWebsiteUpdate(su *mantil.StaticWebsiteUpdate, action mantil.UpdateAction) {
	var idx int = -1
	for i, s := range d.project.StaticWebsites {
		if su.Name == s.Name {
			idx = i
			break
		}
	}
	switch action {
	case mantil.Update:
		if idx == -1 {
			return
		}
		s := d.project.StaticWebsites[idx]
		s.Hash = su.Hash
		d.project.StaticWebsites[idx] = s
	case mantil.Add:
		s := mantil.StaticWebsite{
			Name: su.Name,
		}
		d.project.StaticWebsites = append(d.project.StaticWebsites, s)
	case mantil.Remove:
		if idx == -1 {
			return
		}
		ss := d.project.StaticWebsites
		ss = append(ss[:idx], ss[idx+1:]...)
		d.project.StaticWebsites = ss
	}
}

func (d *Deploy) Deploy() error {
	infrastructureChanged := false
	for _, u := range d.updates {
		if u.Action == mantil.Add || u.Action == mantil.Remove {
			infrastructureChanged = true
		}
	}
	d.updateProject()

	// apply terraform if there are changes in infrastructure
	if infrastructureChanged {
		log.Info("applying changes to infrastructure...")
		if err := d.applyInfrastructure(); err != nil {
			return err
		}
		return mantil.SaveProject(d.project)
	}

	// otherwise just update lambda functions directly
	for _, u := range d.updates {
		fu := u.Function
		if fu == nil {
			continue
		}
		log.Info("updating function %s...", fu.Name)
		if err := d.updateLambdaFunction(fu); err != nil {
			log.Error(err)
		} else {
			log.Info("successfully updated function %s", fu.Name)
		}
	}
	return mantil.SaveProject(d.project)
}

func (d *Deploy) applyInfrastructure() error {
	tf := d.tf
	if err := tf.ApplyForProject(d.project, false); err != nil {
		return fmt.Errorf("could not apply terraform for project %s - %v", d.project.Name, err)
	}
	url, err := tf.Output("url", true)
	if err != nil {
		return fmt.Errorf("could not read terraform output variable for api url - %v", err)
	}
	if url != "" {
		d.project.ApiURL = url
	}
	sites, err := tf.Output("static_websites", false)
	if err != nil {
		return fmt.Errorf("coult not read terraform output variable for static websites - %v", err)
	}
	if err := d.updateWebsitesConfig(sites); err != nil {
		return err
	}
	return nil
}

func (d *Deploy) updateLambdaFunction(f *mantil.FunctionUpdate) error {
	lambdaName := mantil.ProjectResource(d.project.Name, f.Name)
	var err error
	if f.S3Key != "" {
		err = d.aws.UpdateLambdaFunctionCodeFromS3(lambdaName, d.project.Bucket, f.S3Key)
	} else {
		err = fmt.Errorf("could not update lambda function %s due to missing key", lambdaName)
	}
	if err != nil {
		return err
	}
	log.Debug("waiting for function's update status to be successful...")
	return d.aws.WaitLambdaFunctionUpdated(lambdaName)
}

func (d *Deploy) updateWebsitesConfig(tfOutput string) error {
	type sitesOutput struct {
		Name   string `json:"name"`
		Bucket string `json:"bucket"`
	}
	os := &[]sitesOutput{}
	if err := json.Unmarshal([]byte(tfOutput), os); err != nil {
		return err
	}
	for _, o := range *os {
		for i, s := range d.project.StaticWebsites {
			if o.Name == s.Name {
				s.Bucket = o.Bucket
				d.project.StaticWebsites[i] = s
			}
		}
	}
	return nil
}
