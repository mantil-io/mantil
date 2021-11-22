package domain

import "fmt"

type NodeExistsError struct {
	Name string
}

func (e *NodeExistsError) Error() string {
	return fmt.Sprintf("node %s already exists", e.Name)
}

type StageExistsError struct {
	Name string
}

func (e *StageExistsError) Error() string {
	return fmt.Sprintf("stage %s already exists", e.Name)
}

type ProjectNoStagesError struct{}

func (e *ProjectNoStagesError) Error() string {
	return fmt.Sprintf("no stages in project")
}

type NodeNotFoundError struct {
	Name string
}

func (e *NodeNotFoundError) Error() string {
	return fmt.Sprintf("node %s not found", e.Name)
}

type WorkspaceNoNodesError struct{}

func (e *WorkspaceNoNodesError) Error() string {
	return fmt.Sprintf("no nodes in workspace")
}

type ProjectNotFoundError struct{}

func (e *ProjectNotFoundError) Error() string {
	return fmt.Sprintf("project not found")
}

var (
	ErrWorkspaceNotFound = fmt.Errorf("workspace not found")
)

type EvironmentConfigValidationError struct {
	Err error
}

func (e *EvironmentConfigValidationError) Error() string {
	return e.Err.Error()
}
