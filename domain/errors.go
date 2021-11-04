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

type NodeNotFoundError struct {
	Name string
}

func (e *NodeNotFoundError) Error() string {
	return fmt.Sprintf("node %s not found", e.Name)
}

var (
	ErrWorkspaceNotFound = fmt.Errorf("workspace not found")
	ErrProjectNotFound   = fmt.Errorf("no Mantil project found")
)

type EvironmentConfigValidationError struct {
	Err error
}

func (e *EvironmentConfigValidationError) Error() string {
	return e.Err.Error()
}
