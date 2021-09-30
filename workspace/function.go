package workspace

type Function struct {
	Name       string            `yaml:"name"`
	Hash       string            `yaml:"hash"`
	S3Key      string            `yaml:"s3_key"`
	Runtime    string            `yaml:"runtime"`
	Handler    string            `yaml:"handler"`
	MemorySize int               `yaml:"memory_size"`
	Timeout    int               `yaml:"timeout"`
	Path       string            `yaml:"path"`
	Env        map[string]string `yaml:"env"`
}

type FunctionDefaults struct {
	Prefix     string `yaml:"prefix"`
	MemorySize int    `yaml:"memory_size"`
	Timeout    int    `yaml:"timeout"`
}

func (f *Function) SetS3Key(key string) {
	f.S3Key = key
}
