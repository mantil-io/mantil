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

type ErrNameTooLong struct {
	Name      string
	MaxLength int
}

func (e *ErrNameTooLong) Error() string {
	return fmt.Sprintf("the name %s is too long, maximum allowed length is %d", e.Name, e.MaxLength)
}

func (e *ErrNameTooLong) UserMessage() string {
	return validationUserMessage(e)
}

type ErrForbiddenCharacters struct {
	Name              string
	AllowedCharacters string
}

func (e *ErrForbiddenCharacters) Error() string {
	return fmt.Sprintf("the name %s contains forbidden characters, it must only contain the following: %s", e.Name, e.AllowedCharacters)
}

func (e *ErrForbiddenCharacters) UserMessage() string {
	return validationUserMessage(e)
}

const (
	maxNameLength                = 16
	allowedCharactersDescription = "numbers, letters and the special characters - and _"
)

var allowedCharactersRegex = regexp.MustCompile(`^[a-zA-Z0-9\-\_]*$`)

type ValidationError interface {
	Error() string
	UserMessage() string
}

func ValidateName(name string) ValidationError {
	if len(name) > maxNameLength {
		return &ErrNameTooLong{
			Name:      name,
			MaxLength: maxNameLength,
		}
	}
	if !allowedCharactersRegex.MatchString(name) {
		return &ErrForbiddenCharacters{
			Name:              name,
			AllowedCharacters: allowedCharactersDescription,
		}
	}
	return nil
}

func validationUserMessage(err ValidationError) string {
	return fmt.Sprintf("Validation error: %v", err)
}
