package workspace

import (
	"fmt"
	"regexp"
)

var reservedFunctionNames = []string{
	"public",
}

func FunctionNameAvailable(name string) bool {
	for _, rn := range reservedFunctionNames {
		if name == rn {
			return false
		}
	}
	return true
}

type ErrReservedName struct {
	Name string
}

func (e *ErrReservedName) Error() string {
	return fmt.Sprintf("name \"%s\" is reserved", e.Name)
}

const (
	maxNameLength = 16
)

var allowedCharactersRegex = regexp.MustCompile(`^[a-zA-Z0-9\-\_]*$`)

type ValidationError struct {
	Name string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation failed for name %s", e.Name)
}

func newValidationError(name string) *ValidationError {
	return &ValidationError{
		Name: name,
	}
}

func ValidateName(name string) *ValidationError {
	if len(name) > maxNameLength {
		return newValidationError(name)
	}
	if !allowedCharactersRegex.MatchString(name) {
		return newValidationError(name)
	}
	return nil
}
