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

type NodeAlreadyUpToDateError struct {
	Name    string
	Version string
}

func (e *NodeAlreadyUpToDateError) Error() string {
	return fmt.Sprintf("node %s already at version %s", e.Name, e.Version)
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

type EnvironmentConfigValidationError struct {
	Err error
}

func (e *EnvironmentConfigValidationError) Error() string {
	return e.Err.Error()
}

type TokenExpiredError struct{}

func (e *TokenExpiredError) Error() string {
	return fmt.Sprintf("token expired")
}

type SSMPathNotFoundError struct{}

func (e *SSMPathNotFoundError) Error() string {
	return fmt.Sprintf("SSM parameter path not found")
}
