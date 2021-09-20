package mantil

type Function struct {
	Name       string
	Hash       string
	S3Key      string
	Runtime    string
	Handler    string
	MemorySize int
	Timeout    int
	Path       string
	Env        map[string]string
}

type FunctionUpdate struct {
	Name  string
	Hash  string
	S3Key string
}

func (f *Function) SetS3Key(key string) {
	f.S3Key = key
}
