package domain

import (
	"github.com/pkg/errors"
)

type Resource struct {
	Name string
	Hash string
}

type resourceDiff struct {
	added   []string
	removed []string
	updated []string
}

func (d *resourceDiff) infrastructureChanged() bool {
	return len(d.added) > 0 || len(d.removed) > 0
}

func (d *resourceDiff) hasUpdates() bool {
	return d.infrastructureChanged() || len(d.updated) > 0
}

type StageDiff struct {
	functions     resourceDiff
	public        resourceDiff
	configChanged bool
}

func (d *StageDiff) HasUpdates() bool {
	return d.functions.hasUpdates() ||
		d.public.hasUpdates() ||
		d.configChanged
}

func (d *StageDiff) HasFunctionUpdates() bool {
	return d.functions.hasUpdates() ||
		d.configChanged
}

func (d *StageDiff) HasPublicUpdates() bool {
	return d.public.hasUpdates()
}

func (d *StageDiff) InfrastructureChanged() bool {
	return d.functions.infrastructureChanged() ||
		d.public.infrastructureChanged() ||
		d.configChanged
}

func (d *StageDiff) UpdatedFunctions() []string {
	return d.functions.updated
}

func (d *StageDiff) FunctionsAddedUpdatedRemoved() (int, int, int) {
	return len(d.functions.added),
		len(d.functions.updated),
		len(d.functions.removed)
}

func (s *Stage) ApplyChanges(funcs []Resource, publicHash string) (*StageDiff, error) {
	funcDiff, err := s.applyFunctionChanges(funcs)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	publicDiff := s.applyPublicChanges(publicHash)
	configChanged := s.applyConfiguration(s.project.environment)
	return &StageDiff{
		functions:     funcDiff,
		public:        publicDiff,
		configChanged: configChanged,
	}, nil
}

func (s *Stage) applyFunctionChanges(localFuncs []Resource) (resourceDiff, error) {
	var diff resourceDiff
	localFuncNames := resourceNames(localFuncs)
	stageFuncNames := s.FunctionNames()
	diff.added = diffArrays(localFuncNames, stageFuncNames)
	if err := s.AddFunctions(diff.added); err != nil {
		return diff, errors.WithStack(err)
	}
	diff.removed = diffArrays(stageFuncNames, localFuncNames)
	s.RemoveFunctions(diff.removed)
	for _, f := range s.Functions {
		for _, lf := range localFuncs {
			if f.Name == lf.Name && f.Hash != lf.Hash {
				f.SetHash(lf.Hash)
				diff.updated = append(diff.updated, f.Name)
				break
			}
		}
	}
	return diff, nil
}

func (s *Stage) applyPublicChanges(hash string) resourceDiff {
	var rd resourceDiff
	if hash == "" {
		return rd
	}
	if !s.HasPublic() && hash != "" {
		s.Public = &Public{
			Hash: hash,
		}
		rd.added = append(rd.added, "public")
		return rd
	}
	if s.Public.Hash != hash {
		s.Public.Hash = hash
		rd.updated = append(rd.updated, "public")
	}
	return rd
}

func resourceNames(rs []Resource) []string {
	var names []string
	for _, r := range rs {
		names = append(names, r.Name)
	}
	return names
}

// returns a1 - a2
func diffArrays(a1 []string, a2 []string) []string {
	m := make(map[string]bool)
	for _, e := range a2 {
		m[e] = true
	}
	var diff []string
	for _, e := range a1 {
		if m[e] {
			continue
		}
		diff = append(diff, e)
	}
	return diff
}
