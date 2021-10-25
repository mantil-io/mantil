package workspace

import (
	"github.com/mantil-io/mantil/cli/log"
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

func (d *StageDiff) UpdatedPublicSites() []string {
	return d.public.updated
}

func (s *Stage) ApplyChanges(funcs, public []Resource) (*StageDiff, error) {
	funcDiff, err := s.applyFunctionChanges(funcs)
	if err != nil {
		return nil, log.Wrap(err)
	}
	publicDiff := s.applyPublicChanges(public)
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
		return diff, log.Wrap(err)
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

func (s *Stage) applyPublicChanges(localPublic []Resource) resourceDiff {
	var diff resourceDiff
	localSiteNames := resourceNames(localPublic)
	stageSiteNames := s.PublicSiteNames()
	diff.added = diffArrays(localSiteNames, stageSiteNames)
	s.AddPublicSites(diff.added)
	diff.removed = diffArrays(stageSiteNames, localSiteNames)
	s.RemovePublicSites(diff.removed)
	for _, ps := range s.PublicSites() {
		for _, ls := range localPublic {
			if ps.Name == ls.Name && ps.Hash != ls.Hash {
				ps.Hash = ls.Hash
				diff.updated = append(diff.updated, ps.Name)
			}
		}
	}
	return diff
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
