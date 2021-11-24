package log

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func wrap(err error) error {
	return Wrap(err)
}

func errSource() error {
	return Wrap(fmt.Errorf("my-error"), "hide error")
}

func TestWrap(t *testing.T) {

	err := wrap(errSource())
	require.Error(t, err)

	buf := bytes.NewBuffer(nil)
	printStack(buf, err)
	t.Logf("\n%s", buf)

	require.True(t, strings.Contains(buf.String(), "my-error"))

	// //show all causers
	// i := 0
	// for {
	// 	i++
	// 	fmt.Printf("%d %s %T\n", i, err, err)
	// 	c, ok := err.(causer)
	// 	if !ok {
	// 		break
	// 	}
	// 	err = c.Cause()
	// }
}
