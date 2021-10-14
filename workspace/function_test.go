package workspace

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMergeEnv(t *testing.T) {
	type testCase struct {
		initialEnv  map[string]string
		sources     []map[string]string
		expectedEnv map[string]string
		changed     bool
	}
	cases := []testCase{
		{
			initialEnv: map[string]string{},
			sources: []map[string]string{
				{
					"k": "v",
				},
			},
			expectedEnv: map[string]string{
				"k": "v",
			},
			changed: true,
		},
		{
			initialEnv: map[string]string{
				"k": "v",
			},
			sources: []map[string]string{
				{
					"k": "v",
				},
			},
			expectedEnv: map[string]string{
				"k": "v",
			},
			changed: false,
		},
		{
			initialEnv: map[string]string{
				"k": "v",
			},
			sources: []map[string]string{
				{
					"k": "v2",
				},
			},
			expectedEnv: map[string]string{
				"k": "v2",
			},
			changed: true,
		},
		{
			initialEnv: map[string]string{
				"k": "v",
			},
			sources: []map[string]string{
				{
					"k2": "v",
				},
			},
			expectedEnv: map[string]string{
				"k":  "v",
				"k2": "v",
			},
			changed: true,
		},
		{
			initialEnv: map[string]string{
				"k": "v",
			},
			sources: []map[string]string{
				{
					"k": "v",
				},
				{
					"k": "v2",
				},
			},
			expectedEnv: map[string]string{
				"k": "v",
			},
			changed: false,
		},
		{
			initialEnv: map[string]string{
				"k": "v",
			},
			sources: []map[string]string{
				{
					"k": "v",
				},
				{
					"k":  "v2",
					"k2": "v",
				},
				{
					"k2": "v2",
				},
			},
			expectedEnv: map[string]string{
				"k":  "v",
				"k2": "v",
			},
			changed: true,
		},
	}
	for _, c := range cases {
		f := &Function{
			Env: c.initialEnv,
		}
		changed := f.mergeEnv(c.sources...)
		assert.Equal(t, c.expectedEnv, f.Env)
		assert.Equal(t, c.changed, changed)
	}
}
