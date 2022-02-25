package domain_test

import (
	"testing"

	. "github.com/mantil-io/mantil/domain"
	"github.com/stretchr/testify/require"
)

func TestStageChangesWithoutNewResources(t *testing.T) {
	s := initStage(&Stage{}, nil)

	diff, err := s.ApplyChanges(nil, "")

	require.Nil(t, err)
	require.NotNil(t, diff)
	require.False(t, diff.HasUpdates())
	require.False(t, diff.HasFunctionUpdates())
	require.False(t, diff.HasPublicUpdates())
	require.False(t, diff.InfrastructureChanged())
	require.Len(t, diff.UpdatedFunctions(), 0)
	require.Len(t, s.Functions, 0)

	added, updated, removed := diff.FunctionsAddedUpdatedRemoved()
	require.Equal(t, added, 0)
	require.Equal(t, updated, 0)
	require.Equal(t, removed, 0)
}

func TestStageChangesInvalidFunctionName(t *testing.T) {
	s := initStage(&Stage{}, nil)

	diff, err := s.ApplyChanges([]Resource{
		{
			Name: "too-long-name-with-invalid-character*",
			Hash: "hash",
		},
	}, "")

	require.Error(t, err)
	require.Nil(t, diff)
}

func TestStageChangesWithNewFunction(t *testing.T) {
	s := initStage(&Stage{}, nil)

	diff, err := s.ApplyChanges([]Resource{
		{
			Name: "func",
			Hash: "hash",
		},
	}, "")

	require.Nil(t, err)
	require.NotNil(t, diff)
	require.True(t, diff.HasUpdates())
	require.True(t, diff.HasFunctionUpdates())
	require.False(t, diff.HasPublicUpdates())
	require.True(t, diff.InfrastructureChanged())
	require.Len(t, diff.UpdatedFunctions(), 1)
	require.Len(t, s.Functions, 1)
	require.Equal(t, "func", s.Functions[0].Name)
	require.Equal(t, "hash", s.Functions[0].Hash)

	added, updated, removed := diff.FunctionsAddedUpdatedRemoved()
	require.Equal(t, added, 1)
	require.Equal(t, updated, 1)
	require.Equal(t, removed, 0)
}

func TestStageChangesWithUpdatedFunction(t *testing.T) {
	s := initStage(&Stage{
		Functions: []*Function{
			{
				Name: "func",
				Hash: "hash",
			},
		},
	}, nil)

	diff, err := s.ApplyChanges([]Resource{
		{
			Name: "func",
			Hash: "new-hash",
		},
	}, "")

	require.Nil(t, err)
	require.NotNil(t, diff)
	require.True(t, diff.HasUpdates())
	require.True(t, diff.HasFunctionUpdates())
	require.False(t, diff.HasPublicUpdates())
	require.False(t, diff.InfrastructureChanged())
	require.Len(t, diff.UpdatedFunctions(), 1)
	require.Len(t, s.Functions, 1)
	require.Equal(t, "func", s.Functions[0].Name)
	require.Equal(t, "new-hash", s.Functions[0].Hash)

	added, updated, removed := diff.FunctionsAddedUpdatedRemoved()
	require.Equal(t, added, 0)
	require.Equal(t, updated, 1)
	require.Equal(t, removed, 0)
}

func TestStageChangesWithNewAndRemovedFunction(t *testing.T) {
	s := initStage(&Stage{
		Functions: []*Function{
			{
				Name: "func",
				Hash: "hash",
			},
		},
	}, nil)

	diff, err := s.ApplyChanges([]Resource{
		{
			Name: "func2",
			Hash: "hash",
		},
	}, "")

	require.Nil(t, err)
	require.NotNil(t, diff)
	require.True(t, diff.HasUpdates())
	require.True(t, diff.HasFunctionUpdates())
	require.False(t, diff.HasPublicUpdates())
	require.True(t, diff.InfrastructureChanged())
	require.Len(t, diff.UpdatedFunctions(), 1)
	require.Len(t, s.Functions, 1)
	require.Equal(t, "func2", s.Functions[0].Name)
	require.Equal(t, "hash", s.Functions[0].Hash)

	added, updated, removed := diff.FunctionsAddedUpdatedRemoved()
	require.Equal(t, added, 1)
	require.Equal(t, updated, 1)
	require.Equal(t, removed, 1)
}

func TestStageChangesWithNewAndUnchangedFunction(t *testing.T) {
	s := initStage(&Stage{
		Functions: []*Function{
			{
				Name: "func",
				Hash: "hash",
			},
		},
	}, nil)

	diff, err := s.ApplyChanges([]Resource{
		{
			Name: "func",
			Hash: "hash",
		},
		{
			Name: "func2",
			Hash: "hash",
		},
	}, "")

	require.Nil(t, err)
	require.NotNil(t, diff)
	require.True(t, diff.HasUpdates())
	require.True(t, diff.HasFunctionUpdates())
	require.False(t, diff.HasPublicUpdates())
	require.True(t, diff.InfrastructureChanged())
	require.Len(t, diff.UpdatedFunctions(), 1)
	require.Len(t, s.Functions, 2)
	require.Equal(t, "func", s.Functions[0].Name)
	require.Equal(t, "hash", s.Functions[0].Hash)
	require.Equal(t, "func2", s.Functions[1].Name)
	require.Equal(t, "hash", s.Functions[1].Hash)

	added, updated, removed := diff.FunctionsAddedUpdatedRemoved()
	require.Equal(t, added, 1)
	require.Equal(t, updated, 1)
	require.Equal(t, removed, 0)
}

func TestStageChangesWithNewFunctionAndPublic(t *testing.T) {
	s := initStage(&Stage{}, nil)

	diff, err := s.ApplyChanges([]Resource{
		{
			Name: "func",
			Hash: "hash",
		},
	}, "hash")

	require.Nil(t, err)
	require.NotNil(t, diff)
	require.True(t, diff.HasUpdates())
	require.True(t, diff.HasFunctionUpdates())
	require.True(t, diff.HasPublicUpdates())
	require.True(t, diff.InfrastructureChanged())
	require.Len(t, diff.UpdatedFunctions(), 1)
	require.Len(t, s.Functions, 1)
	require.Equal(t, "func", s.Functions[0].Name)
	require.Equal(t, "hash", s.Functions[0].Hash)

	added, updated, removed := diff.FunctionsAddedUpdatedRemoved()
	require.Equal(t, added, 1)
	require.Equal(t, updated, 1)
	require.Equal(t, removed, 0)

}

func TestStageChangesWithUpdatedPublic(t *testing.T) {
	s := initStage(&Stage{
		Public: &Public{
			Hash: "hash",
		},
	}, nil)

	diff, err := s.ApplyChanges(nil, "new-hash")

	require.Nil(t, err)
	require.NotNil(t, diff)
	require.True(t, diff.HasUpdates())
	require.False(t, diff.HasFunctionUpdates())
	require.True(t, diff.HasPublicUpdates())
	require.False(t, diff.InfrastructureChanged())
	require.Len(t, diff.UpdatedFunctions(), 0)
	require.Equal(t, "new-hash", s.Public.Hash)

	added, updated, removed := diff.FunctionsAddedUpdatedRemoved()
	require.Equal(t, added, 0)
	require.Equal(t, updated, 0)
	require.Equal(t, removed, 0)
}

func TestStageChangesWithProjectConfiguration(t *testing.T) {
	s := initStage(&Stage{
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

	diff, err := s.ApplyChanges([]Resource{
		{
			Name: "func",
			Hash: "hash",
		},
	}, "")

	require.Nil(t, err)
	require.NotNil(t, diff)
	require.True(t, diff.HasUpdates())
	require.True(t, diff.HasFunctionUpdates())
	require.False(t, diff.HasPublicUpdates())
	require.True(t, diff.InfrastructureChanged())
	require.Len(t, diff.UpdatedFunctions(), 0)
	require.Len(t, s.Functions, 1)
	require.Equal(t, 128, s.Functions[0].MemorySize)
	require.Equal(t, "v", s.Functions[0].Env["k"])

	added, updated, removed := diff.FunctionsAddedUpdatedRemoved()
	require.Equal(t, added, 0)
	require.Equal(t, updated, 0)
	require.Equal(t, removed, 0)
}

func TestStageChangesWithStageConfiguration(t *testing.T) {
	s := initStage(&Stage{
		Functions: []*Function{
			{
				Name: "func",
				Hash: "hash",
			},
		},
	}, &EnvironmentConfig{
		Project: ProjectEnvironmentConfig{
			FunctionConfiguration: FunctionConfiguration{
				MemorySize: 512,
				Timeout:    60,
				Env: map[string]string{
					"k":  "v2",
					"k2": "v",
				},
			},
			Stages: []StageEnvironmentConfig{
				{
					FunctionConfiguration: FunctionConfiguration{
						MemorySize: 128,
						Env: map[string]string{
							"k": "v",
						},
					},
				},
				{
					Name: "wrong-stage",
					FunctionConfiguration: FunctionConfiguration{
						MemorySize: 256,
						Env: map[string]string{
							"k4": "v4",
						},
					},
				},
			},
		},
	})

	diff, err := s.ApplyChanges([]Resource{
		{
			Name: "func",
			Hash: "hash",
		},
	}, "")

	require.Nil(t, err)
	require.NotNil(t, diff)
	require.True(t, diff.HasUpdates())
	require.True(t, diff.HasFunctionUpdates())
	require.False(t, diff.HasPublicUpdates())
	require.True(t, diff.InfrastructureChanged())
	require.Len(t, diff.UpdatedFunctions(), 0)
	require.Len(t, s.Functions, 1)
	require.Equal(t, 128, s.Functions[0].MemorySize)
	require.Equal(t, 60, s.Functions[0].Timeout)
	require.Equal(t, "v", s.Functions[0].Env["k"])
	require.Equal(t, "v", s.Functions[0].Env["k2"])

	added, updated, removed := diff.FunctionsAddedUpdatedRemoved()
	require.Equal(t, added, 0)
	require.Equal(t, updated, 0)
	require.Equal(t, removed, 0)
}

func TestStageChangesWithFunctionConfiguration(t *testing.T) {
	s := initStage(&Stage{
		Functions: []*Function{
			{
				Name: "func",
				Hash: "hash",
			},
		},
	}, &EnvironmentConfig{
		Project: ProjectEnvironmentConfig{
			FunctionConfiguration: FunctionConfiguration{
				MemorySize: 512,
				Timeout:    60,
				Env: map[string]string{
					"k":  "v2",
					"k2": "v",
				},
			},
			Stages: []StageEnvironmentConfig{
				{
					FunctionConfiguration: FunctionConfiguration{
						MemorySize: 128,
						Env: map[string]string{
							"k": "v",
						},
					},
					Functions: []FunctionEnvironmentConfig{
						{
							Name: "func",
							FunctionConfiguration: FunctionConfiguration{
								MemorySize: 64,
								Timeout:    900,
								Env: map[string]string{
									"k3": "v3",
									"k":  "v3",
								},
							},
						},
						{
							Name: "func2",
							FunctionConfiguration: FunctionConfiguration{
								MemorySize: 256,
								Timeout:    600,
								Env: map[string]string{
									"k4": "v4",
								},
							},
						},
					},
				},
			},
		},
	})

	diff, err := s.ApplyChanges([]Resource{
		{
			Name: "func",
			Hash: "hash",
		},
		{
			Name: "func2",
			Hash: "hash,",
		},
	}, "")

	require.Nil(t, err)
	require.NotNil(t, diff)
	require.True(t, diff.HasUpdates())
	require.True(t, diff.HasFunctionUpdates())
	require.False(t, diff.HasPublicUpdates())
	require.True(t, diff.InfrastructureChanged())
	require.Len(t, diff.UpdatedFunctions(), 1)
	require.Len(t, s.Functions, 2)
	require.Equal(t, 64, s.Functions[0].MemorySize)
	require.Equal(t, 900, s.Functions[0].Timeout)
	require.Equal(t, "v3", s.Functions[0].Env["k"])
	require.Equal(t, "v", s.Functions[0].Env["k2"])
	require.Equal(t, "v3", s.Functions[0].Env["k3"])
	require.Equal(t, 256, s.Functions[1].MemorySize)
	require.Equal(t, 600, s.Functions[1].Timeout)
	require.Equal(t, "v", s.Functions[1].Env["k"])
	require.Equal(t, "v", s.Functions[1].Env["k2"])
	require.Equal(t, "v4", s.Functions[1].Env["k4"])

	added, updated, removed := diff.FunctionsAddedUpdatedRemoved()
	require.Equal(t, added, 1)
	require.Equal(t, updated, 1)
	require.Equal(t, removed, 0)
}

func TestStageChangesDefaultConfiguration(t *testing.T) {
	s := initStage(&Stage{
		Name:    "stage",
		Default: true,
		Functions: []*Function{
			{
				Name: "func",
				Hash: "hash",
			},
		},
	}, &EnvironmentConfig{
		Project: ProjectEnvironmentConfig{
			Stages: []StageEnvironmentConfig{
				{
					Name: "stage",
					CustomDomain: CustomDomain{
						DomainName: "domain",
					},
				},
			},
		}},
	)

	diff, err := s.ApplyChanges([]Resource{
		{
			Name: "func",
			Hash: "hash",
		},
	}, "")

	require.Nil(t, err)
	require.NotNil(t, diff)
	require.True(t, diff.HasUpdates())
	require.True(t, diff.HasFunctionUpdates())
	require.False(t, diff.HasPublicUpdates())
	require.True(t, diff.InfrastructureChanged())
	require.Len(t, diff.UpdatedFunctions(), 0)
	require.Len(t, s.Functions, 1)
	require.Equal(t, 128, s.Functions[0].MemorySize)
	require.Equal(t, 900, s.Functions[0].Timeout)
	require.Equal(t, s.Project().DefaultStage().Node().Name, s.Functions[0].Env[EnvWorkspace])
	require.Equal(t, s.Project().DefaultStage().Node().ID, s.Functions[0].Env[EnvKey])
	require.Equal(t, s.Project().Name, s.Functions[0].Env[EnvProjectName])
	require.Equal(t, s.Name, s.Functions[0].Env[EnvStageName])

	added, updated, removed := diff.FunctionsAddedUpdatedRemoved()
	require.Equal(t, added, 0)
	require.Equal(t, updated, 0)
	require.Equal(t, removed, 0)
}

func initStage(s *Stage, env *EnvironmentConfig) *Stage {
	workspace := Workspace{
		Nodes: []*Node{{
			ID:   "uid",
			Name: s.NodeName}},
	}
	project := Project{
		Name:   "project",
		Stages: []*Stage{s},
	}
	Factory(&workspace, &project, env)
	if s.Public == nil {
		s.Public = &Public{}
	}
	return s
}
