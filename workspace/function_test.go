package workspace

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMergeConfiguration(t *testing.T) {
	type testCase struct {
		initialConfig  FunctionConfiguration
		sources        []FunctionConfiguration
		expectedConfig FunctionConfiguration
		changed        bool
	}
	cases := []testCase{
		{
			initialConfig: FunctionConfiguration{},
			sources: []FunctionConfiguration{
				{
					MemorySize: 128,
					Timeout:    30,
				},
			},
			expectedConfig: FunctionConfiguration{
				MemorySize: 128,
				Timeout:    30,
			},
			changed: true,
		},
		{
			initialConfig: FunctionConfiguration{
				MemorySize: 128,
				Timeout:    30,
			},
			sources: []FunctionConfiguration{
				{
					MemorySize: 128,
					Timeout:    30,
				},
			},
			expectedConfig: FunctionConfiguration{
				MemorySize: 128,
				Timeout:    30,
			},
			changed: false,
		},
		{
			initialConfig: FunctionConfiguration{},
			sources: []FunctionConfiguration{
				{
					MemorySize: 128,
					Timeout:    30,
				},
				{
					MemorySize: 512,
					Timeout:    60,
				},
			},
			expectedConfig: FunctionConfiguration{
				MemorySize: 512,
				Timeout:    60,
			},
			changed: true,
		},
		{
			initialConfig: FunctionConfiguration{},
			sources: []FunctionConfiguration{
				{
					MemorySize: 512,
					Timeout:    60,
				},
				{
					MemorySize: 128,
					Timeout:    30,
				},
			},
			expectedConfig: FunctionConfiguration{
				MemorySize: 128,
				Timeout:    30,
			},
			changed: true,
		},
		{
			initialConfig: FunctionConfiguration{
				MemorySize: 512,
				Timeout:    60,
			},
			sources: []FunctionConfiguration{
				{
					MemorySize: 512,
					Timeout:    60,
				},
				{
					MemorySize: 128,
					Timeout:    30,
				},
			},
			expectedConfig: FunctionConfiguration{
				MemorySize: 128,
				Timeout:    30,
			},
			changed: true,
		},
		{
			initialConfig: FunctionConfiguration{
				MemorySize: 512,
				Timeout:    60,
			},
			sources: []FunctionConfiguration{
				{
					MemorySize: 512,
					Timeout:    60,
				},
				{
					MemorySize: 128,
					Timeout:    30,
				},
			},
			expectedConfig: FunctionConfiguration{
				MemorySize: 128,
				Timeout:    30,
			},
			changed: true,
		},
		{
			initialConfig: FunctionConfiguration{
				MemorySize: 512,
			},
			sources: []FunctionConfiguration{
				{
					Timeout: 60,
				},
			},
			expectedConfig: FunctionConfiguration{
				MemorySize: 0,
				Timeout:    60,
			},
			changed: true,
		},
		{
			initialConfig: FunctionConfiguration{
				Env: map[string]string{},
			},
			sources: []FunctionConfiguration{
				{
					Env: map[string]string{
						"k": "v",
					},
				},
			},
			expectedConfig: FunctionConfiguration{
				Env: map[string]string{
					"k": "v",
				},
			},
			changed: true,
		},
		{
			initialConfig: FunctionConfiguration{
				Env: map[string]string{},
			},
			sources: []FunctionConfiguration{
				{
					MemorySize: 128,
					Timeout:    30,
				},
				{
					Env: map[string]string{
						"k": "v",
					},
				},
			},
			expectedConfig: FunctionConfiguration{
				MemorySize: 128,
				Timeout:    30,
				Env: map[string]string{
					"k": "v",
				},
			},
			changed: true,
		},
		{
			initialConfig: FunctionConfiguration{
				Env: map[string]string{},
			},
			sources: []FunctionConfiguration{
				{
					Env: map[string]string{
						"k": "v",
					},
				},
			},
			expectedConfig: FunctionConfiguration{
				Env: map[string]string{
					"k": "v",
				},
			},
			changed: true,
		},
		{
			initialConfig: FunctionConfiguration{
				Env: map[string]string{
					"k": "v",
				},
			},
			sources: []FunctionConfiguration{
				{
					Env: map[string]string{
						"k": "v",
					},
				},
			},
			expectedConfig: FunctionConfiguration{
				Env: map[string]string{
					"k": "v",
				},
			},
			changed: false,
		},
		{
			initialConfig: FunctionConfiguration{
				Env: map[string]string{
					"k": "v",
				},
			},
			sources: []FunctionConfiguration{
				{
					Env: map[string]string{
						"k": "v2",
					},
				},
			},
			expectedConfig: FunctionConfiguration{
				Env: map[string]string{
					"k": "v2",
				},
			},
			changed: true,
		},
		{
			initialConfig: FunctionConfiguration{
				Env: map[string]string{
					"k": "v",
				},
			},
			sources: []FunctionConfiguration{
				{
					Env: map[string]string{
						"k2": "v",
					},
				},
			},
			expectedConfig: FunctionConfiguration{
				Env: map[string]string{
					"k2": "v",
				},
			},
			changed: true,
		},
		{
			initialConfig: FunctionConfiguration{
				Env: map[string]string{
					"k": "v",
				},
			},
			sources: []FunctionConfiguration{
				{
					Env: map[string]string{
						"k": "v",
					},
				},
				{
					Env: map[string]string{
						"k": "v2",
					},
				},
			},
			expectedConfig: FunctionConfiguration{
				Env: map[string]string{
					"k": "v2",
				},
			},
			changed: true,
		},
		{
			initialConfig: FunctionConfiguration{
				Env: map[string]string{
					"k": "v",
				},
			},
			sources: []FunctionConfiguration{
				{
					Env: map[string]string{
						"k": "v",
					},
				},
				{
					Env: map[string]string{
						"k":  "v2",
						"k2": "v",
					},
				},
				{
					Env: map[string]string{
						"k2": "v2",
					},
				},
			},
			expectedConfig: FunctionConfiguration{
				Env: map[string]string{
					"k":  "v2",
					"k2": "v2",
				},
			},
			changed: true,
		},
	}
	for _, c := range cases {
		fc := c.initialConfig
		changed := fc.merge(c.sources...)
		assert.Equal(t, c.expectedConfig, fc)
		assert.Equal(t, c.changed, changed)
	}
}
