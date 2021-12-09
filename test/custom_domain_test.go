package test

import (
	"io/ioutil"
	"path/filepath"
	"testing"

	"github.com/mantil-io/mantil/domain"
	"gopkg.in/yaml.v2"
)

func TestCustomDomain(t *testing.T) {
	c := newClitestWithWorkspaceCopy(t)
	t.Parallel()

	projectName := "custom-domain"
	c.Run("mantil", "new", projectName).Success()
	c.Cd(projectName)

	c.WithWorkdir(func() {
		createCustomDomainConfig(t, domain.CustomDomain{
			DomainName: "unit-test.mantil.team",
		})
	})
	c.Run("mantil", "stage", "new", "stage", "--node", defaultNodeName).Success().
		Contains("Endpoint: https://api.unit-test.mantil.team")
	c.Run("mantil", "invoke", "ping").Success().Contains("pong")
	c.Run("curl", "https://api.unit-test.mantil.team/ping/").Success().Contains("pong")
	c.WithWorkdir(func() {
		createCustomDomainConfig(t, domain.CustomDomain{
			DomainName:    "unit-test.mantil.team",
			HttpSubdomain: "http",
		})
	})
	c.Run("mantil", "deploy").Success()
	c.Run("mantil", "invoke", "ping").Success().Contains("pong")
	c.Run("curl", "https://api.unit-test.mantil.team/ping/").Fail()
	c.Run("curl", "https://http.unit-test.mantil.team/ping/").Success().Contains("pong")

	c.Run("mantil", "stage", "destroy", "--all", "--yes").Success()
}

func createCustomDomainConfig(t *testing.T, cd domain.CustomDomain) {
	ec := domain.EnvironmentConfig{
		Project: domain.ProjectEnvironmentConfig{
			Stages: []domain.StageEnvironmentConfig{
				{
					Name:         "stage",
					CustomDomain: cd,
				},
			},
		},
	}
	buf, err := yaml.Marshal(ec)
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join("config", "environment.yml")
	if err := ioutil.WriteFile(path, buf, 0644); err != nil {
		t.Fatal(err)
	}
}
