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

type ErrForbiddenCharacters struct {
	Name              string
	AllowedCharacters string
}

func (e *ErrForbiddenCharacters) Error() string {
	return fmt.Sprintf("the name %s contains forbidden characters, it must only contain the following: %s", e.Name, e.AllowedCharacters)
}

const (
	maxNameLength                = 16
	allowedCharactersDescription = "numbers, letters and the special character -"
)

var allowedCharactersRegex = regexp.MustCompile(`^[a-zA-Z0-9\-]+$`)

func ValidateName(name string) error {
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
