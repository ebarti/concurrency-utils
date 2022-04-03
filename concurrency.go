package utils

// Bridge gets all values coming from a stream of channels and streams them on a single channel
func Bridge[T any, V ~chan T](done chan T, chanStream chan V) chan T {
	valStream := make(chan T)
	go func() {
		defer close(valStream)
		for {
			var stream chan T
			select {
			case maybeStream, ok := <-chanStream:
				if ok == false {
					return
				}
				stream = maybeStream
			case <-done:
				return
			}
			for val := range OrDone[T](done, stream) {
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
func OrDone[T any](done, c chan T) chan T {
	valStream := make(chan T)
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

// Tee de-multiplexes a channel into multiple receivers
func Tee[T any](in chan T, receivers ...chan T) {
	go func() {
		defer func() {
			for i := range receivers {
				close(receivers[i])
			}
		}()
		for elem := range in {
			for ii := 0; ii < len(receivers); ii++ {
				receivers[ii] <- elem
			}
		}
	}()
}

// TeeValue de-multiplexes a value into multiple receivers
func TeeValue[T any](in T, receivers ...chan T) {
	for ii := 0; ii < len(receivers); ii++ {
		receivers[ii] <- in
	}
}

// Take is a generator utility and best used in combination with Repeat
// It reads a given number of values from a valueStream, but returns
// early if a value is placed in the done channel.
func Take[S, T any](done chan S, valueStream chan T, num int) chan T {
	takeStream := make(chan T)
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
func Repeat[T any](done <-chan interface{}, v []T) <-chan T {
	valueStream := make(chan T)
	go func() {
		defer close(valueStream)
		for {
			for _, val := range v {
				select {
				case <-done:
					return
				case valueStream <- val:
				}
			}
		}
	}()
	return valueStream
}
