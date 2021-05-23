package concurrencyutils

import (
	"reflect"
	"testing"
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
	bridge := Bridge(nil, genVals())
	var vals []interface{}
	for v := range bridge {
		vals = append(vals, v)
	}
	if len(vals) != 10 {
		t.Errorf("Bridge test failed. Expected 10 values but received only %d", len(vals))
	}
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

	c1, c2, c3, c4 := make(chan interface{}), make(chan interface{}), make(chan interface{}), make(chan interface{})
	var c1v, c2v, c3v, c4v []interface{}
	Tee(genVals(), c1, c2, c3, c4)
	for c := range c1 {
		c1v = append(c1v, c)
		c2v = append(c2v, <-c2)
		c3v = append(c3v, <-c3)
		c4v = append(c4v, <-c4)
	}
	if len(c1v) != 10 {
		t.Errorf("Tee test failed. Channel 1 expected 10 values but received only %d", len(c1v))
	}
	if len(c2v) != 10 {
		t.Errorf("Tee test failed. Channel 2 expected 10 values but received only %d", len(c2v))
	}
	if len(c3v) != 10 {
		t.Errorf("Tee test failed. Channel 3 expected 10 values but received only %d", len(c3v))
	}
	if len(c4v) != 10 {
		t.Errorf("Tee test failed. Channel 4 expected 10 values but received only %d", len(c4v))
	}
}

func TestOrDone(t *testing.T) {
	type args struct {
		sendDoneAtIdx int
		c             []interface{}
	}
	tests := []struct {
		name string
		args args
		want []interface{}
	}{
		{
			name: "Test never done",
			args: args{
				sendDoneAtIdx: -1,
				c:             []interface{}{0, 1, 2, 3, 4, 5, 6, 7},
			},
			want: []interface{}{0, 1, 2, 3, 4, 5, 6, 7},
		},
		{
			name: "Test done in between",
			args: args{
				sendDoneAtIdx: 4,
				c:             []interface{}{0, 1, 2, 3, 4, 5, 6, 7},
			},
			want: []interface{}{0, 1, 2, 3, 4},
		},
		{
			name: "Test done",
			args: args{
				sendDoneAtIdx: 0,
				c:             []interface{}{0, 1, 2, 3, 4, 5, 6, 7},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			done := make(chan interface{})
			c := make(chan interface{})
			go func() {
				defer func() {
					close(done)
					close(c)
				}()
				for i, v := range tt.args.c {
					if i == tt.args.sendDoneAtIdx {
						done <- 1
						return
					}
					c <- v
				}
			}()
			got := OrDone(done, c)
			if tt.want != nil {
				i := 0
				for gotValue := range got {
					if !reflect.DeepEqual(gotValue, tt.want[i]) {
						t.Errorf("OrDone() = got %v, but wanted %v", gotValue, tt.want[i])
					}
					i++
				}
			} else {
				if len(got) > 0 {
					t.Error("Got values but shouldn't have")
				}
			}
		})
	}
}

func TestTake(t *testing.T) {
	valueStreamFn := func(doneAtIdx int, value interface{}) (_, _ <-chan interface{}) {
		valueStream := make(chan interface{}, 1)
		doneStream := make(chan interface{}, 1)

		go func() {
			defer func() {
				close(valueStream)
				close(doneStream)
			}()
			i := 0
			for {
				if i == doneAtIdx {
					for len(valueStream) > 0 {
						// block
					}
					doneStream <- 1
					break
				}
				valueStream <- value
				i++
			}
		}()
		return valueStream, doneStream
	}
	type args struct {
		sendDoneAtIdx int
		value         interface{}
		num           int
	}
	tests := []struct {
		name string
		args args
		want []interface{}
	}{
		{
			name: "When done takes no values",
			args: args{
				sendDoneAtIdx: 0,
				value:         0,
				num:           10,
			},
			want: nil,
		},
		{
			name: "When never done takes all values",
			args: args{
				sendDoneAtIdx: -1,
				value:         0,
				num:           10,
			},
			want: []interface{}{0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		},
		{
			name: "When done in the middle takes only values up to done",
			args: args{
				sendDoneAtIdx: 4,
				value:         0,
				num:           10,
			},
			want: []interface{}{0, 0, 0, 0, 0},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			valueStream, doneStream := valueStreamFn(tt.args.sendDoneAtIdx, tt.args.value)
			got := Take(doneStream, valueStream, tt.args.num)
			if tt.want != nil {
				i := 0
				for gotValue := range got {
					if !reflect.DeepEqual(gotValue, tt.want[i]) {
						t.Errorf("Take() = got %v, but wanted %v", gotValue, tt.want[i])
					}
					i++
				}
			} else {
				if len(got) > 0 {
					t.Error("Take failed: Got values but shouldn't have")
				}
			}
		})
	}
}
