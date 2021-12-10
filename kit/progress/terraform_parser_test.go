package progress

import (
	"bufio"
	"io/ioutil"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseLog(t *testing.T) {
	p := NewTerraformParser()
	testStateChanges(t, p, "testdata/terraform_apply_output.txt", map[int]ParserState{
		1:   StateInitializing,
		765: StateCreating,
		852: StateDone,
	})
	require.Nil(t, p.Error())
	require.Len(t, p.Outputs, 4)
	require.Equal(t, "mantil-aef7a9da", p.Outputs["functions_bucket"])
	require.Equal(t, "", p.Outputs["public_site_bucket"])
	require.Equal(t, "https://9mosxdgpy2.execute-api.eu-central-1.amazonaws.com", p.Outputs["url"])
	require.Equal(t, "wss://976orve3jg.execute-api.eu-central-1.amazonaws.com", p.Outputs["ws_url"])

	p = NewTerraformParser()
	testStateChanges(t, p, "testdata/terraform_destroy_output.txt", map[int]ParserState{
		1:    StateInitializing,
		1384: StateDestroying,
		1455: StateDone,
	})
	require.Nil(t, p.Error())

	p = NewTerraformParser()
	testStateChanges(t, p, "testdata/terraform_error_output.txt", map[int]ParserState{
		1:   StateInitial,
		690: StateUpdating,
		701: StateDone,
	})
	require.NotNil(t, p.Error())
}

func testStateChanges(t *testing.T, p *TerraformParser, dataPath string, stateChanges map[int]ParserState) {
	content, err := ioutil.ReadFile(dataPath)
	require.NoError(t, err)
	require.Equal(t, StateInitial, p.State())

	scanner := bufio.NewScanner(strings.NewReader(string(content)))
	currentState := StateInitial
	lineCnt := 0
	for scanner.Scan() {
		lineCnt++
		line := scanner.Text()
		isTf := p.Parse(line)
		require.True(t, isTf, line)
		s, ok := stateChanges[lineCnt]
		if ok {
			currentState = s
		}
		require.Equal(t, p.State(), currentState)
	}
}
