package log

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestStore(t *testing.T) {
	t.Run("store", func(t *testing.T) {
		es := newEventsStore()
		es.push([]byte("pero"))
		es.push([]byte("zdero"))
		require.NoError(t, es.store())
	})

	t.Run("restore", func(t *testing.T) {
		es := newEventsStore()
		require.NoError(t, es.restore())
		require.Len(t, es.events, 2)
		for _, v := range es.events {
			s := string(v)
			if s != "pero" && s != "zdero" {
				t.Errorf("must be one of these")
			}
		}

		t.Run("clear", func(t *testing.T) {
			require.NoError(t, es.clear())
			require.Len(t, es.events, 0)
		})
	})

	t.Run("clear with nothing on dis", func(t *testing.T) {
		es := newEventsStore()
		es.push([]byte("pero"))
		es.push([]byte("zdero"))
		require.NoError(t, es.clear())
	})

}
