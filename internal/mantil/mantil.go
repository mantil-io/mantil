package mantil

import (
	"fmt"
)

func RuntimeResource(v ...string) string {
	r := "mantil"
	for _, n := range v {
		r = fmt.Sprintf("%s-%s", r, n)
	}
	return r
}
