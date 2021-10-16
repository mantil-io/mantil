package workspace

import (
	"fmt"
)

// TODO remove
// used in ws, set env variable of that function to know the name of other lambda function
func RuntimeResource(v ...string) string {
	r := "mantil"
	for _, n := range v {
		r = fmt.Sprintf("%s-%s", r, n)
	}
	return r
}
