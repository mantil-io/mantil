package workspace

import (
	"fmt"
	"reflect"
)

type Function struct {
	Name                  string `yaml:"name"`
	Hash                  string `yaml:"hash"`
	S3Key                 string `yaml:"s3_key"`
	FunctionConfiguration `yaml:",inline"`
	stage                 *Stage
}

func (f *Function) SetHash(hash string) {
	f.Hash = hash
	f.S3Key = fmt.Sprintf("%s/functions/%s-%s.zip", f.stage.BucketPrefix(), f.Name, f.Hash)
}

func (f *Function) LambdaName() string {
	return LambdaFunctionName(
		f.stage.project.Name,
		f.stage.Name,
		f.Name,
		f.stage.account.ResourceSuffix(),
	)
}

// TODO: global out of domain model
func LambdaFunctionName(projectName, stageName, functionName, suffix string) string {
	return fmt.Sprintf("%s-%s-%s-%s",
		projectName,
		stageName,
		functionName,
		suffix,
	)
}

func (f *Function) addDefaults() {
	if f.MemorySize == 0 {
		f.MemorySize = 128
	}
	if f.Timeout == 0 {
		f.Timeout = 60 * 15
	}
	if f.Env == nil {
		f.Env = make(map[string]string)
	}
}

type FunctionConfiguration struct {
	MemorySize int               `yaml:"memory_size" jsonschema:"minimum=128,maximum=10240"`
	Timeout    int               `yaml:"timeout" jsonschema:"minimum=1,maximum=900"`
	Env        map[string]string `yaml:"env"`
}

// merge function configuration from multiple sources ordered by priority
// from lowest to highest, returns true if any changes have occurred
func (fc *FunctionConfiguration) merge(sources ...FunctionConfiguration) bool {
	merged := FunctionConfiguration{}
	for _, s := range sources {
		if s.MemorySize != 0 {
			merged.MemorySize = s.MemorySize
		}
		if s.Timeout != 0 {
			merged.Timeout = s.Timeout
		}
		for k, v := range s.Env {
			if merged.Env == nil {
				merged.Env = make(map[string]string)
			}
			merged.Env[k] = v
		}
	}
	changed := merged.changed(fc)
	*fc = merged
	return changed
}

func (fc *FunctionConfiguration) changed(original *FunctionConfiguration) bool {
	return !reflect.DeepEqual(fc, original)
}
