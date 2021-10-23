package workspace_test

import (
	"testing"

	. "github.com/mantil-io/mantil/workspace"
	"github.com/stretchr/testify/require"
)

func TestApplyChanges(t *testing.T) {
	s := initStage(&Stage{}, nil)
	diff, err := s.ApplyChanges(nil, nil)
	require.Nil(t, err)
	require.NotNil(t, diff)
	require.False(t, diff.HasUpdates())
	require.False(t, diff.HasFunctionUpdates())
	require.False(t, diff.HasPublicUpdates())
	require.False(t, diff.InfrastructureChanged())
	require.Len(t, diff.UpdatedFunctions(), 0)
	require.Len(t, diff.UpdatedPublicSites(), 0)
	require.Len(t, s.Functions, 0)
	require.Len(t, s.Public.Sites, 0)

	s = initStage(&Stage{}, nil)
	diff, err = s.ApplyChanges([]Resource{
		{
			Name: "func",
			Hash: "hash",
		},
	}, nil)
	require.Nil(t, err)
	require.NotNil(t, diff)
	require.True(t, diff.HasUpdates())
	require.True(t, diff.HasFunctionUpdates())
	require.False(t, diff.HasPublicUpdates())
	require.True(t, diff.InfrastructureChanged())
	require.Len(t, diff.UpdatedFunctions(), 1)
	require.Len(t, diff.UpdatedPublicSites(), 0)
	require.Len(t, s.Functions, 1)
	require.Equal(t, "func", s.Functions[0].Name)
	require.Equal(t, "hash", s.Functions[0].Hash)
	require.Len(t, s.Public.Sites, 0)

	s = initStage(&Stage{
		Functions: []*Function{
			{
				Name: "func",
				Hash: "hash",
			},
		},
	}, nil)
	diff, err = s.ApplyChanges([]Resource{
		{
			Name: "func",
			Hash: "new-hash",
		},
	}, nil)
	require.Nil(t, err)
	require.NotNil(t, diff)
	require.True(t, diff.HasUpdates())
	require.True(t, diff.HasFunctionUpdates())
	require.False(t, diff.HasPublicUpdates())
	require.False(t, diff.InfrastructureChanged())
	require.Len(t, diff.UpdatedFunctions(), 1)
	require.Len(t, diff.UpdatedPublicSites(), 0)
	require.Len(t, s.Functions, 1)
	require.Equal(t, "func", s.Functions[0].Name)
	require.Equal(t, "new-hash", s.Functions[0].Hash)
	require.Len(t, s.Public.Sites, 0)

	s = initStage(&Stage{
		Functions: []*Function{
			{
				Name: "func",
				Hash: "hash",
			},
		},
	}, nil)
	diff, err = s.ApplyChanges([]Resource{
		{
			Name: "func2",
			Hash: "hash",
		},
	}, nil)
	require.Nil(t, err)
	require.NotNil(t, diff)
	require.True(t, diff.HasUpdates())
	require.True(t, diff.HasFunctionUpdates())
	require.False(t, diff.HasPublicUpdates())
	require.True(t, diff.InfrastructureChanged())
	require.Len(t, diff.UpdatedFunctions(), 1)
	require.Len(t, diff.UpdatedPublicSites(), 0)
	require.Len(t, s.Functions, 1)
	require.Equal(t, "func2", s.Functions[0].Name)
	require.Equal(t, "hash", s.Functions[0].Hash)
	require.Len(t, s.Public.Sites, 0)

	s = initStage(&Stage{
		Functions: []*Function{
			{
				Name: "func",
				Hash: "hash",
			},
		},
	}, nil)
	diff, err = s.ApplyChanges([]Resource{
		{
			Name: "func",
			Hash: "hash",
		},
		{
			Name: "func2",
			Hash: "hash",
		},
	}, nil)
	require.Nil(t, err)
	require.NotNil(t, diff)
	require.True(t, diff.HasUpdates())
	require.True(t, diff.HasFunctionUpdates())
	require.False(t, diff.HasPublicUpdates())
	require.True(t, diff.InfrastructureChanged())
	require.Len(t, diff.UpdatedFunctions(), 1)
	require.Len(t, diff.UpdatedPublicSites(), 0)
	require.Len(t, s.Functions, 2)
	require.Equal(t, "func", s.Functions[0].Name)
	require.Equal(t, "hash", s.Functions[0].Hash)
	require.Equal(t, "func2", s.Functions[1].Name)
	require.Equal(t, "hash", s.Functions[1].Hash)
	require.Len(t, s.Public.Sites, 0)

	s = initStage(&Stage{}, nil)
	diff, err = s.ApplyChanges([]Resource{
		{
			Name: "func",
			Hash: "hash",
		},
	}, []Resource{
		{
			Name: "public",
			Hash: "hash",
		},
	})
	require.Nil(t, err)
	require.NotNil(t, diff)
	require.True(t, diff.HasUpdates())
	require.True(t, diff.HasFunctionUpdates())
	require.True(t, diff.HasPublicUpdates())
	require.True(t, diff.InfrastructureChanged())
	require.Len(t, diff.UpdatedFunctions(), 1)
	require.Len(t, diff.UpdatedPublicSites(), 1)
	require.Len(t, s.Functions, 1)
	require.Equal(t, "func", s.Functions[0].Name)
	require.Equal(t, "hash", s.Functions[0].Hash)
	require.Len(t, s.Public.Sites, 1)
	require.Equal(t, "public", s.Public.Sites[0].Name)
	require.Equal(t, "hash", s.Public.Sites[0].Hash)

	s = initStage(&Stage{}, &EnvironmentConfig{
		Project: ProjectEnvironmentConfig{
			FunctionConfiguration: FunctionConfiguration{
				MemorySize: 128,
				Env: map[string]string{
					"k": "v",
				},
			},
		},
	})
	diff, err = s.ApplyChanges(nil, nil)
	require.Nil(t, err)
	require.NotNil(t, diff)
	require.False(t, diff.HasUpdates())
	require.False(t, diff.HasFunctionUpdates())
	require.False(t, diff.HasPublicUpdates())
	require.False(t, diff.InfrastructureChanged())
	require.Len(t, diff.UpdatedFunctions(), 0)
	require.Len(t, diff.UpdatedPublicSites(), 0)
	require.Len(t, s.Functions, 0)
	require.Len(t, s.Public.Sites, 0)

	s = initStage(&Stage{
		Functions: []*Function{
			{
				Name: "func",
				Hash: "hash",
			},
		},
	}, &EnvironmentConfig{
		Project: ProjectEnvironmentConfig{
			FunctionConfiguration: FunctionConfiguration{
				MemorySize: 128,
				Env: map[string]string{
					"k": "v",
				},
			},
		},
	})
	diff, err = s.ApplyChanges([]Resource{
		{
			Name: "func",
			Hash: "hash",
		},
	}, nil)
	require.Nil(t, err)
	require.NotNil(t, diff)
	require.True(t, diff.HasUpdates())
	require.True(t, diff.HasFunctionUpdates())
	require.False(t, diff.HasPublicUpdates())
	require.True(t, diff.InfrastructureChanged())
	require.Len(t, diff.UpdatedFunctions(), 0)
	require.Len(t, diff.UpdatedPublicSites(), 0)
	require.Len(t, s.Functions, 1)
	require.Equal(t, 128, s.Functions[0].MemorySize)
	require.Equal(t, "v", s.Functions[0].Env["k"])
	require.Len(t, s.Public.Sites, 0)
}

func initStage(s *Stage, env *EnvironmentConfig) *Stage {
	workspace := Workspace{
		Accounts: []*Account{{Name: s.AccountName}},
	}
	project := Project{
		Stages: []*Stage{s},
	}
	Factory(&workspace, &project, env)
	if s.Public == nil {
		s.Public = &Public{}
	}
	return s
}
