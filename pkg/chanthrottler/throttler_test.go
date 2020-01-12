package chanthrottler

import "testing"

import "time"

import "reflect"

func TestThrottler(t *testing.T) {
	type readVal struct {
		val interface{}
		ok  bool
	}
	tests := []struct {
		name string
		f    func(chan interface{})
		want readVal
	}{
		{
			name: "only push latest value",
			f: func(c chan interface{}) {
				for i := 1; i <= 1024; i++ {
					c <- i
				}
			},
			want: readVal{1024, true},
		},
		{
			name: "only close if that is the latest action",
			f: func(c chan interface{}) {
				for i := 1; i <= 1024; i++ {
					c <- i
				}
				close(c)
			},
			want: readVal{nil, false},
		},
	}
	for _, v := range tests {
		t.Run(v.name, func(tt *testing.T) {
			src := make(chan interface{})
			dst := Throttle(32*time.Millisecond, src)
			v.f(src)
			val, ok := <-dst
			got := readVal{val, ok}
			if !reflect.DeepEqual(v.want, got) {
				t.Errorf("Channel Recv: want %v, got %v", v.want, got)
			}
		})
	}
}
