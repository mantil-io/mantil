package main

import (
	"math/rand"
	"time"

	"github.com/mantil-io/mantil/kit/progress"
)

func main() {
	c := progress.NewCounter(42)
	p := progress.New(
		"Waiting",
		progress.LogFunc,
		c,
		progress.NewDots(),
	)
	p.Run()
	for i := 1; i <= 42; i++ {
		time.Sleep(time.Duration(rand.Intn(1000)) * time.Millisecond)
		c.SetCount(i)
	}
	p.Done()
}
