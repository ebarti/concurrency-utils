package concurrencyutils

import (
	"testing"
	"time"
)

func TestBridge(t *testing.T) {
	genVals := func() <-chan <-chan interface{} {
		chanStream := make(chan (<-chan interface{}))
		go func() {
			defer close(chanStream)
			for i := 0; i < 10; i++ {
				stream := make(chan interface{}, 1)
				stream <- i
				close(stream)
				chanStream <- stream
			}
		}()
		return chanStream
	}
	_ = genVals()
}

func TestOrDone(t *testing.T) {
	sig := func(idx int, after time.Duration) <-chan interface{} {
		c := make(chan interface{})
		go func() {
			defer close(c)
			c <- idx
			time.Sleep(after)
		}()
		return c
	}
	_ = sig(1, 1 * time.Hour)
}

func TestTee(t *testing.T) {
	genVals := func() <-chan interface{} {
		retChan := make(chan interface{})
		go func() {
			defer close(retChan)
			for i := 0; i < 10; i++ {
				retChan <- i
			}
		}()
		return retChan
	}
	c1, c2 := Tee(nil, genVals())
	var c1v, c2v []interface{}
	for c := range c1 {
		c1v = append(c1v, c)
		c2v = append(c2v, <-c2)
	}

	if len(c1v) != 10 {
		t.Error("Not enough outs")
	}
	if len(c2v) != 10 {
		t.Error("Not enough outs")
	}
}

