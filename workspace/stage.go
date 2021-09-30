package workspace

const (
	DefaultStageName = "dev"
)

type Stage struct {
	Name        string            `yaml:"name"`
	Default     bool              `yaml:"default,omitempty"`
	Account     string            `yaml:"account"`
	Endpoints   *StageEndpoints   `yaml:"endpoints,omitempty"`
	Env         map[string]string `yaml:"env,omitempty"`
	Functions   []*Function       `yaml:"functions,omitempty"`
	PublicSites []*PublicSite     `yaml:"public_sites,omitempty"`
}

type StageEndpoints struct {
	Rest string `yaml:"rest"`
	Ws   string `yaml:"ws"`
}
