package workspace

type PublicSite struct {
	Name   string `yaml:"name"`
	Bucket string `yaml:"bucket"`
	Hash   string `yaml:"hash"`
}
