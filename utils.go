package concurrencyutils

import (
	"reflect"
)

// Bridge gets all values coming from a stream of channels and streams them on a single channel
func Bridge(done <-chan interface{}, chanStream <-chan <-chan interface{}) <-chan interface{} {
	valStream := make(chan interface{})
	go func() {
		defer close(valStream)
		for {
			var stream <-chan interface{}
			select {
			case maybeStream, ok := <-chanStream:
				if ok == false {
					return
				}
				stream = maybeStream
			case <-done:
				return
			}
			for val := range OrDone(done, stream) {
				select {
				case valStream <- val:
				case <-done:
				}
			}
		}
	}()
	return valStream
}

// OrDone gets either reads from the input channel or returns if a
// value is fed to the done channel
func OrDone(done, c <-chan interface{}) <-chan interface{} {
	valStream := make(chan interface{})
	go func() {
		defer close(valStream)
		for {
			select {
			case <-done:
				return
			case v, ok := <-c:
				if ok == false {
					return
				}
				select {
				case valStream <- v:
				case <-done:
				}
			}
		}
	}()
	return valStream
}

// Tee demultiplexes a channel into multiple receivers
func Tee(in <-chan interface{}, receivers ...chan<- interface{}) {
	cases := make([]reflect.SelectCase, len(receivers))
	go func() {
		defer func() {
			for i := range receivers {
				close(receivers[i])
			}
		}()
		for i := range cases {
			cases[i].Dir = reflect.SelectSend
		}
		for elem := range in {
			for i := range cases {
				cases[i].Chan = reflect.ValueOf(receivers[i])
				cases[i].Send = reflect.ValueOf(elem)
			}
			for range cases {
				chosen, _, _ := reflect.Select(cases)
				cases[chosen].Chan = reflect.ValueOf(nil)
			}
		}
	}()
}

// TeeValue demultiplexes a channel into multiple receivers
func TeeValue(in interface{}, receivers ...chan<- interface{}) {
	cases := make([]reflect.SelectCase, len(receivers))
	for i := range cases {
		cases[i].Dir = reflect.SelectSend
	}
	for i := range cases {
		cases[i].Chan = reflect.ValueOf(receivers[i])
		cases[i].Send = reflect.ValueOf(in)
	}
	for range cases {
		chosen, _, _ := reflect.Select(cases)
		cases[chosen].Chan = reflect.ValueOf(nil)
	}
}

// Take is a generator utility and best used in combination with Repeat
// It reads a given number of values from a valueStream, but returns
// early if a value is placed in the done channel.
func Take(done <-chan interface{}, valueStream <-chan interface{}, num int) <-chan interface{} {
	takeStream := make(chan interface{})
	go func() {
		defer close(takeStream)
		for i := 0; i < num; i++ {
			select {
			case <-done:
				return
			case in := <-valueStream:
				takeStream <- in
			}
		}
	}()
	return takeStream
}

// Repeat will repeat the values you pass to it infinitely until a value is placed in the done channel.
func Repeat(done <-chan interface{}, values ...interface{}) <-chan interface{} {
	valueStream := make(chan interface{})
	go func() {
		defer close(valueStream)
		for {
			for _, v := range values {
				select {
				case <-done:
					return
				case valueStream <- v:
				}
			}
		}
	}()
	return valueStream
}
