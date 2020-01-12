package chanthrottler

import (
	"reflect"
	"time"
)

func recvWithTimeout(duration time.Duration, v reflect.Value) (x reflect.Value, ok, timedOut bool) {
	chosen, val, ok := reflect.Select([]reflect.SelectCase{
		{
			Dir:  reflect.SelectRecv,
			Chan: v,
		},
		{
			Dir:  reflect.SelectRecv,
			Chan: reflect.ValueOf(time.After(duration)),
		},
	})
	switch chosen {
	case 0:
		return val, ok, false
	case 1:
		return reflect.Value{}, false, true
	default:
		panic("uncovered case")
	}
}

// Throttle throttles the passed channel. It drops all messages except for
// the one that takes longer than duration since last message.
func Throttle(duration time.Duration, ch interface{}) <-chan interface{} {
	v := reflect.ValueOf(ch)
	if v.IsNil() {
		panic("ch was nil")
	}
	if v.Kind() != reflect.Chan {
		panic("ch was not of kind chan")
	}

	c := make(chan interface{})
	go func() {
		val, ok := v.Recv()
		for {
			if !ok {
				close(c)
				return
			}
			nval, nok, to := recvWithTimeout(duration, v)
			if to {
				c <- val.Interface()
				return
			}
			val, ok = nval, nok
		}
	}()
	return c
}
