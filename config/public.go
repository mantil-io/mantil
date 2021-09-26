package config

type PublicSite struct {
	Name   string `yaml:"name"`
	Bucket string `yaml:"bucket"`
	Hash   string `yaml:"hash"`
}

type PublicSiteUpdate struct {
	Name string
	Hash string
}