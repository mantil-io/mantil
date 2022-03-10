package controller

import (
	"sort"

	"github.com/mantil-io/mantil/cli/log"
	"github.com/mantil-io/mantil/cli/ui"
	"github.com/mantil-io/mantil/domain"
)

func Nodes() error {
	fs, err := newStore()
	if err != nil {
		return log.Wrap(err)
	}
	nodes, err := fs.Workspace().NodeList()
	if err != nil {
		return err
	}
	if len(nodes) == 0 {
		return log.Wrap(&domain.WorkspaceNoNodesError{})
	}
	var data [][]string
	for _, n := range nodes {
		data = append(data, []string{n.Name, n.AccountID, n.Region, n.ID, n.Version})
	}
	ShowTable([]string{"name", "AWS Account", "AWS Region", "ID", "Version"}, data)
	return nil
}

type AwsResourcesArgs struct {
	Nodes bool
	Stage string
}

func (a AwsResourcesArgs) empty() bool {
	return a.Stage == "" && !a.Nodes
}

func NewAwsResources(a AwsResourcesArgs) *AwsResources {
	return &AwsResources{
		args: a,
	}
}

type AwsResources struct {
	args AwsResourcesArgs
}

func (a *AwsResources) Show() error {
	if a.args.empty() {
		if err := a.showStage(a.args.Stage); err == nil {
			return nil
		}
		return a.showNodes()
	}

	if a.args.Stage != "" {
		if err := a.showStage(a.args.Stage); err != nil {
			return err
		}
	}
	if a.args.Nodes {
		if err := a.showNodes(); err != nil {
			return err
		}
	}
	return nil
}

func (a *AwsResources) showStage(stageName string) error {
	_, stage, err := newStoreWithStage(stageName)
	if err != nil {
		return err
	}
	a.stage(stage)
	return nil
}

func (a *AwsResources) showNodes() error {
	store, err := newStore()
	if err != nil {
		return err
	}
	for _, node := range store.Workspace().Nodes {
		a.node(node)
	}
	return nil
}

func (a *AwsResources) project() error {
	return nil
}

func (a *AwsResources) stage(st *domain.Stage) {
	ui.Title("\nProject %s stage %s\n", st.Project().Name, st.Name)
	ui.Info("Resources:")
	a.showResourcesTable(st.Resources())
	ui.Info("Tags:")
	a.showTagsTable(st.ResourceTags())

	a.node(st.Node())
}

func (a *AwsResources) showResourcesTable(rs []domain.AwsResource) {
	var data [][]string
	for _, rs := range rs {
		data = append(data, []string{rs.Name, rs.Type, rs.AWSName, rs.LogGroup()})
	}
	ShowTable([]string{"name", "type", "AWS resource name", "cloudwatch log group"}, data)
}

func (a *AwsResources) showTagsTable(tgs map[string]string) {
	var tags [][]string
	for k, v := range tgs {
		tags = append(tags, []string{k, v})
	}
	sort.Slice(tags, func(i, j int) bool {
		return tags[i][0] < tags[j][0]
	})
	ShowTable([]string{"key", "value"}, tags)
}

func (a *AwsResources) node(n *domain.Node) {
	ui.Title("\nNode %s\n", n.Name)
	ui.Info("Resources:")
	a.showResourcesTable(n.Resources())
	ui.Info("Tags:")
	a.showTagsTable(n.ResourceTags())
}
