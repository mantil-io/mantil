package workspace

type Public struct {
	Bucket string        `yaml:"bucket"`
	Sites  []*PublicSite `yaml:"sites"`
}

type PublicSite struct {
	Name string `yaml:"name"`
	Hash string `yaml:"hash"`
}
