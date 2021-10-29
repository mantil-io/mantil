package domain

import "fmt"

type AccountExistsError struct {
	Name string
}

func (e *AccountExistsError) Error() string {
	return fmt.Sprintf("account %s already exists", e.Name)
}

type StageExistsError struct {
	Name string
}

func (e *StageExistsError) Error() string {
	return fmt.Sprintf("stage %s already exists", e.Name)
}

type AccountNotFoundError struct {
	Name string
}

func (e *AccountNotFoundError) Error() string {
	return fmt.Sprintf("account %s not found", e.Name)
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
