package domain

import (
	"fmt"
	"regexp"
)

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
