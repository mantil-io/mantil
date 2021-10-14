package workspace

import "fmt"

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

func (f *Function) addDefaults() {
	if f.Path == "" {
		f.Path = f.Name
	}
	if f.S3Key == "" {
		if f.Hash != "" {
			f.S3Key = fmt.Sprintf("functions/%s-%s.zip", f.Name, f.Hash)
		} else {
			f.S3Key = fmt.Sprintf("functions/%s.zip", f.Name)
		}
	}
	if f.Runtime == "" {
		f.Runtime = "provided.al2"
	}
	if f.MemorySize == 0 {
		f.MemorySize = 128
	}
	if f.Timeout == 0 {
		f.Timeout = 60 * 15
	}
	if f.Handler == "" {
		f.Handler = "bootstrap"
	}
	if f.Env == nil {
		f.Env = make(map[string]string)
	}
}

// merge environment variables from multiple sources
// which are ordered by priority, from highest to lowest
func (f *Function) mergeEnv(sources ...map[string]string) bool {
	// gather all keys
	keys := make(map[string]bool)
	for _, s := range sources {
		for k := range s {
			keys[k] = true
		}
	}
	changed := false
	for k := range keys {
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
	// remove old variables
	for k := range f.Env {
		if _, ok := keys[k]; !ok {
			delete(f.Env, k)
			changed = true
		}
	}
	return changed
}
