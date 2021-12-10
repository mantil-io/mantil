package progress

import (
	"fmt"
	"strings"
	"sync"
	"time"
)

type Element interface {
	UpdateChan() <-chan struct{}
	Current() string
	Stop()
	TerminalOnly() bool
}

type ticker interface {
	C() <-chan time.Time
	Stop()
}

type standardTicker struct {
	ticker *time.Ticker
}

func (st *standardTicker) C() <-chan time.Time {
	return st.ticker.C
}

func (st *standardTicker) Stop() {
	st.ticker.Stop()
}

type Dots struct {
	currentCnt int
	ticker     ticker
	updateCh   chan struct{}
	done       chan struct{}
	closer     sync.Once
}

func NewDots() *Dots {
	return newDots(
		&standardTicker{
			ticker: time.NewTicker(time.Second),
		},
	)
}

func newDots(ticker ticker) *Dots {
	d := &Dots{
		ticker:   ticker,
		updateCh: make(chan struct{}),
		done:     make(chan struct{}),
	}
	go d.loop()
	return d
}

func (d *Dots) loop() {
	for {
		select {
		case <-d.ticker.C():
			d.currentCnt = (d.currentCnt + 1) % 4
			d.updateCh <- struct{}{}
		case <-d.done:
			d.ticker.Stop()
			close(d.updateCh)
			return
		}
	}
}

func (d *Dots) Stop() {
	d.closer.Do(func() {
		close(d.done)
	})
}

func (d *Dots) UpdateChan() <-chan struct{} {
	return d.updateCh
}

func (d *Dots) Current() string {
	if d.isDone() {
		return ""
	}
	dots := strings.Repeat(".", d.currentCnt)
	return fmt.Sprintf("%-4s", dots)
}

func (d *Dots) TerminalOnly() bool {
	return true
}

func (d *Dots) isDone() bool {
	select {
	case <-d.done:
		return true
	default:
		return false
	}
}

type Counter struct {
	total    int
	current  int
	updateCh chan struct{}
}

func NewCounter(total int) *Counter {
	c := &Counter{
		total:    total,
		updateCh: make(chan struct{}),
	}
	return c
}

func (c *Counter) SetCount(value int) {
	if value == c.current {
		return
	}
	c.current = value
	c.updateCh <- struct{}{}
}

func (c *Counter) Current() string {
	cur := fmt.Sprintf(" %d%% (%d/%d)",
		int(100*float64(c.current)/float64(c.total)),
		c.current,
		c.total,
	)
	cur = strings.ReplaceAll(cur, "%", "%%")
	return cur
}

func (c *Counter) UpdateChan() <-chan struct{} {
	return c.updateCh
}

func (c *Counter) Stop() {}

func (c *Counter) TerminalOnly() bool {
	return false
}
