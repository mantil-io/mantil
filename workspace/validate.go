package workspace

import "fmt"

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
