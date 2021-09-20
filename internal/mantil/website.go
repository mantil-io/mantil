package mantil

type StaticWebsite struct {
	Name   string
	Bucket string
	Hash   string
}

type StaticWebsiteUpdate struct {
	Name string
	Hash string
}
