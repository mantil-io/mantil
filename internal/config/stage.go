package config

const (
	DefaultStageName = "dev"
)

type Stage struct {
	Name        string            `yaml:"name"`
	Account     string            `yaml:"account"`
	Endpoints   *StageEndpoints   `yaml:"endpoints"`
	Env         map[string]string `yaml:"env"`
	Functions   []*Function       `yaml:"functions"`
	PublicSites []*PublicSite     `yaml:"public_sites"`
}

type StageEndpoints struct {
	Rest string `yaml:"rest"`
	Ws   string `yaml:"ws"`
}
