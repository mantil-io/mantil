package cmd

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGenerateMd(t *testing.T) {
	cmd := root()
	var g mdGenerator
	h := g.help(cmd)
	fmt.Printf("%s", h)

}

func TestReplaceForReadme(t *testing.T) {
	m1 := regexp.MustCompile(`mantil_user_register\.md.\s*\|\s*(.*)\s*\|`)

	text := "Neki moj text"
	str := "| [mantil user register](mantil_user_register.md) | todo |"
	expected := "| [mantil user register](mantil_user_register.md) | Neki moj text |"

	rst := m1.FindAllStringSubmatch(str, 1)
	fmt.Printf("rst '%v'\n", rst[0][1])

	actual := m1.ReplaceAllString(str, text)

	require.Equal(t, expected, actual)
}
