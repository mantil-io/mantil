package deploy

import (
	"fmt"
	"log"

	"github.com/atoz-technology/mantil-backend/internal/assets"
	"github.com/atoz-technology/mantil-backend/internal/aws"
	"github.com/atoz-technology/mantil-backend/internal/mantil"
	"github.com/atoz-technology/mantil-backend/internal/terraform"
)

const (
	FunctionsDir = "functions"
	BuildDir     = "build"
)

type Deploy struct {
	aws             *aws.AWS
	project         *mantil.Project
	functionUpdates []mantil.FunctionUpdate
	path            string
}

func New(project *mantil.Project, updates []mantil.FunctionUpdate, path string) (*Deploy, error) {
	awsClient, err := aws.New()
	if err != nil {
		return nil, err
	}
	assets.StartServer()
	return &Deploy{
		aws:             awsClient,
		project:         project,
		functionUpdates: updates,
		path:            path,
	}, nil
}

func (d *Deploy) updateProject() {
	for _, fu := range d.functionUpdates {
		var idx int = -1
		for i, f := range d.project.Functions {
			if fu.Name == f.Name {
				idx = i
				break
			}
		}
		switch {
		case fu.Updated:
			if idx == -1 {
				continue
			}
			f := d.project.Functions[idx]
			f.Hash = fu.Hash
			f.S3Key = fu.S3Key
			f.ImageKey = fu.ImageKey
			d.project.Functions[idx] = f
		case fu.Added:
			f := mantil.Function{
				Name:     fu.Name,
				Hash:     fu.Hash,
				S3Key:    fu.S3Key,
				ImageKey: fu.ImageKey,
			}
			d.project.Functions = append(d.project.Functions, f)
		case fu.Removed:
			if idx == -1 {
				continue
			}
			fs := d.project.Functions
			fs = append(fs[:idx], fs[idx+1:]...)
			d.project.Functions = fs
		}
	}
	d.project.AddFunctionDefaults()
}

func (d *Deploy) Deploy() error {
	infrastructureChanged := false
	for _, u := range d.functionUpdates {
		if u.Added || u.Removed {
			infrastructureChanged = true
		}
	}
	d.updateProject()

	// if there are changes in infrastructure let terraform update all the necessary functions among other changes
	if infrastructureChanged {
		log.Printf("applying terraform due to infrastructure changes")
		if err := d.applyInfrastructure(); err != nil {
			return err
		}
		return mantil.SaveProject(d.project)
	}

	// otherwise just update lambda functions directly
	for _, u := range d.functionUpdates {
		log.Printf("updating function %s...", u.Name)
		if err := d.updateLambdaFunction(u); err != nil {
			log.Print(err)
		} else {
			log.Printf("successfully updated function %s", u.Name)
		}
	}
	return mantil.SaveProject(d.project)
}

func (d *Deploy) applyInfrastructure() error {
	tf := terraform.New(d.path)
	if err := tf.ApplyForProject(d.project, false); err != nil {
		return fmt.Errorf("could not apply terraform for project %s - %v", d.project.Name, err)
	}
	url, err := tf.Output("url")
	if err != nil {
		return fmt.Errorf("could not read terraform output variable for api url - %v", err)
	}
	if url != "" {
		d.project.ApiURL = url
	}
	return nil
}

func (d *Deploy) updateLambdaFunction(f mantil.FunctionUpdate) error {
	lambdaName := fmt.Sprintf("%s-mantil-team-%s-%s", d.project.Organization.Name, d.project.Name, f.Name)
	var err error
	if f.S3Key != "" {
		err = d.aws.UpdateLambdaFunctionCodeFromS3(lambdaName, d.project.Bucket, f.S3Key)
	} else if f.ImageKey != "" {
		err = d.aws.UpdateLambdaFunctionCodeImage(lambdaName, f.ImageKey)
	} else {
		err = fmt.Errorf("could not update lambda function %s due to missing key", lambdaName)
	}
	if err != nil {
		return err
	}
	log.Printf("waiting for function's update status to be successful...")
	return d.aws.WaitLambdaFunctionUpdated(lambdaName)
}
