package config

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
