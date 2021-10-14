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

// merge environment variables from multiple sources
// which are ordered by priority, from highest to lowest
func (f *Function) mergeEnv(sources ...map[string]string) bool {
	// gather all keys
	keysMap := make(map[string]bool)
	for _, s := range sources {
		for k := range s {
			keysMap[k] = true
		}
	}
	var keys []string
	for k := range keysMap {
		keys = append(keys, k)
	}
	changed := false
	for _, k := range keys {
		// apply values according to priority
		for _, s := range sources {
			v, ok := s[k]
			if !ok {
				continue
			}
			if f.Env[k] == v {
				break
			}
			f.Env[k] = v
			changed = true
			break
		}
	}
	return changed
}
