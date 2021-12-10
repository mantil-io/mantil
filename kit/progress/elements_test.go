package progress

import (
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type testTicker struct {
	ch chan time.Time
}

func newTestTicker() *testTicker {
	return &testTicker{
		ch: make(chan time.Time),
	}
}

func (t *testTicker) C() <-chan time.Time {
	return t.ch
}

func (t *testTicker) tick() {
	t.ch <- time.Now()
}

func (t *testTicker) Stop() {
	close(t.ch)
}

func TestDots(t *testing.T) {
	ticker := newTestTicker()
	d := newDots(ticker)

	require.Equal(t, "    ", d.Current())
	ticker.tick()
	<-d.UpdateChan()
	require.Equal(t, ".   ", d.Current())
	ticker.tick()
	<-d.UpdateChan()
	require.Equal(t, "..  ", d.Current())
	ticker.tick()
	<-d.UpdateChan()
	require.Equal(t, "... ", d.Current())
	ticker.tick()
	<-d.UpdateChan()
	require.Equal(t, "    ", d.Current())

	d.Stop()
	require.Equal(t, "", d.Current())
}

func TestCounter(t *testing.T) {
	c := NewCounter(100)
	require.Equal(t, " 0%% (0/100)", c.Current())
	for i := 1; i <= 100; i++ {
		go func() {
			c.SetCount(i)
		}()
		<-c.UpdateChan()
		iStr := strconv.Itoa(i)
		expected := " " + iStr + "%% (" + iStr + "/100)"
		require.Equal(t, expected, c.Current())
	}
}
