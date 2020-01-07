package rprovider

import (
	"fmt"
	"reflect"
	"testing"
	"time"
)

type ArgsNone struct{}
type ArgsStr struct{ first string }
type ArgsErr struct{ first error }
type ArgsReaderValueErr struct {
	first  ReaderValue
	second error
}

type AdderDouble struct {
	addCalls []ArgsStr
	addRets  []ArgsErr
}

func (w *AdderDouble) Add(path string) error {
	if len(w.addRets) == 0 {
		panic("Too many calls to Add()")
	}
	w.addCalls = append(w.addCalls, ArgsStr{path})
	ret := w.addRets[0]
	w.addRets = w.addRets[1:]
	return ret.first
}

type CloserDouble struct {
	closeCalls []ArgsNone
	closeRets  []ArgsErr
}

func (w *CloserDouble) Close() error {
	if len(w.closeRets) == 0 {
		panic("Too many calls to Close()")
	}
	w.closeCalls = append(w.closeCalls, ArgsNone{})
	ret := w.closeRets[0]
	w.closeRets = w.closeRets[1:]
	return ret.first
}

type WatcherDouble struct {
	*AdderDouble
	*CloserDouble
	ch chan WatcherMsg
}

func (w *WatcherDouble) Chan() <-chan WatcherMsg {
	return w.ch
}

func NewWatcherDoubleFactoryFunc(addRets []ArgsErr, closeRets []ArgsErr) WatcherFactoryFunc {
	d := &WatcherDouble{
		AdderDouble: &AdderDouble{
			addRets: addRets,
		},
		CloserDouble: &CloserDouble{
			closeRets: closeRets,
		},
		ch: make(chan WatcherMsg),
	}
	return func() (Watcher, error) {
		return d, nil
	}
}

type ReaderDouble struct {
	*AdderDouble
	*CloserDouble
	readCalls []ArgsNone
	readRets  []ArgsReaderValueErr
}

func (w *ReaderDouble) Read() (ReaderValue, error) {
	if len(w.readRets) == 0 {
		panic("Too many calls to Read()")
	}
	w.readCalls = append(w.readCalls, struct{}{})
	ret := w.readRets[0]
	w.readRets = w.readRets[1:]
	return ret.first, ret.second
}

func NewReaderDoubleFactoryFunc(addRets []ArgsErr, closeRets []ArgsErr, readRets []ArgsReaderValueErr) ReaderFactoryFunc {
	d := &ReaderDouble{
		AdderDouble: &AdderDouble{
			addRets: addRets,
		},
		CloserDouble: &CloserDouble{
			closeRets: closeRets,
		},
		readRets: readRets,
	}
	return func() (Reader, error) {
		return d, nil
	}
}

func TestNew(t *testing.T) {
	type args struct {
		readerFactory  ReaderFactoryFunc
		watcherFactory WatcherFactoryFunc
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "readerFactory errors",
			args: args{
				readerFactory: func() (Reader, error) {
					return nil, fmt.Errorf("foo")
				},
				watcherFactory: func() (Watcher, error) {
					return &WatcherDouble{}, nil
				},
			},
			wantErr: true,
		},
		{
			name: "watcherFactory errors",
			args: args{
				readerFactory: func() (Reader, error) {
					return &ReaderDouble{}, nil
				},
				watcherFactory: func() (Watcher, error) {
					return nil, fmt.Errorf("foo")
				},
			},
			wantErr: true,
		},
		{
			name: "nil readerFactory",
			args: args{
				readerFactory: nil,
				watcherFactory: func() (Watcher, error) {
					return nil, fmt.Errorf("foo")
				},
			},
			wantErr: true,
		},
		{
			name: "nil watcherFactory",
			args: args{
				readerFactory: func() (Reader, error) {
					return &ReaderDouble{}, nil
				},
				watcherFactory: nil,
			},
			wantErr: true,
		},
		{
			name: "does not error normally",
			args: args{
				readerFactory: func() (Reader, error) {
					return &ReaderDouble{}, nil
				},
				watcherFactory: func() (Watcher, error) {
					return &WatcherDouble{}, nil
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := New(tt.args.readerFactory, tt.args.watcherFactory)
			if (err != nil) != tt.wantErr {
				t.Errorf("New() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func TestProvider_Channel(t *testing.T) {
	t.Run("returns non-nil channel", func(t *testing.T) {
		p, err := New(
			func() (Reader, error) { return &ReaderDouble{}, nil },
			func() (Watcher, error) { return &WatcherDouble{}, nil },
		)
		if err != nil {
			t.Errorf("New() error %v", err)
		}
		c := p.Channel()
		if c == nil {
			t.Error("Provider.Channel() returned nil")
		}
	})
}

func TestProvider_Close(t *testing.T) {
	tests := []struct {
		name             string
		readerFactory    ReaderFactoryFunc
		watcherFactory   WatcherFactoryFunc
		wantChannelClose bool
		wantErr          bool
	}{
		{
			name:             "closes channel",
			readerFactory:    NewReaderDoubleFactoryFunc([]ArgsErr{}, []ArgsErr{ArgsErr{nil}}, []ArgsReaderValueErr{}),
			watcherFactory:   NewWatcherDoubleFactoryFunc([]ArgsErr{}, []ArgsErr{ArgsErr{nil}}),
			wantChannelClose: true,
			wantErr:          false,
		},
		{
			name:             "closes channel and returns error when reader error occurs",
			readerFactory:    NewReaderDoubleFactoryFunc([]ArgsErr{}, []ArgsErr{ArgsErr{fmt.Errorf("foo")}}, []ArgsReaderValueErr{}),
			watcherFactory:   NewWatcherDoubleFactoryFunc([]ArgsErr{}, []ArgsErr{ArgsErr{nil}}),
			wantChannelClose: true,
			wantErr:          true,
		},
		{
			name:             "closes channel and returns error when watcher error occurs",
			readerFactory:    NewReaderDoubleFactoryFunc([]ArgsErr{}, []ArgsErr{ArgsErr{nil}}, []ArgsReaderValueErr{}),
			watcherFactory:   NewWatcherDoubleFactoryFunc([]ArgsErr{}, []ArgsErr{ArgsErr{fmt.Errorf("foo")}}),
			wantChannelClose: true,
			wantErr:          true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l, err := New(tt.readerFactory, tt.watcherFactory)
			if err != nil {
				t.Errorf("New() error %v", err)
			}
			c := l.Channel()

			if err := l.Close(); (err != nil) != tt.wantErr {
				t.Errorf("Provider.Close() error = %v, wantErr %v", err, tt.wantErr)
			}
			ok := true
			select {
			case _, ok = <-c:
			case <-time.After(100 * time.Millisecond):
				t.Fatal("Provider.Close() did not close channel in time")
			}

			if ok == tt.wantChannelClose {
				t.Errorf("Provider.Close() Channel Closed got %v, want %v", !ok, tt.wantChannelClose)
			}
		})
	}
}

func TestProvider_Add(t *testing.T) {
	tests := []struct {
		name           string
		watcherAddRets []ArgsErr
		readerAddRets  []ArgsErr
		addCalls       []ArgsStr
		wantErrs       []bool
	}{
		{
			name: "calls add on watcher and reader",
			watcherAddRets: []ArgsErr{
				ArgsErr{nil},
				ArgsErr{nil},
			},
			readerAddRets: []ArgsErr{
				ArgsErr{nil},
				ArgsErr{nil},
			},
			addCalls: []ArgsStr{
				ArgsStr{"foo"},
				ArgsStr{"bar"},
			},
			wantErrs: []bool{false, false},
		},
		{
			name:           "returns err when Watcher.Add() errors",
			watcherAddRets: []ArgsErr{ArgsErr{fmt.Errorf("foo")}},
			readerAddRets:  []ArgsErr{ArgsErr{nil}},
			addCalls:       []ArgsStr{ArgsStr{"foo"}},
			wantErrs:       []bool{true},
		},
		{
			name:           "returns err when Reader.Add() errors",
			watcherAddRets: []ArgsErr{ArgsErr{nil}},
			readerAddRets:  []ArgsErr{ArgsErr{fmt.Errorf("foo")}},
			addCalls:       []ArgsStr{ArgsStr{"foo"}},
			wantErrs:       []bool{true},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var rve []ArgsReaderValueErr
			for i := 0; i < len(tt.addCalls); i++ {
				rve = append(rve, ArgsReaderValueErr{nil, nil})
			}
			rf := NewReaderDoubleFactoryFunc(tt.readerAddRets, []ArgsErr{}, rve)
			wf := NewWatcherDoubleFactoryFunc(tt.watcherAddRets, []ArgsErr{})
			ri, _ := rf()
			wi, _ := wf()
			r := ri.(*ReaderDouble)
			w := wi.(*WatcherDouble)

			ec := make(chan struct{})
			l, err := New(rf, wf)
			if err != nil {
				t.Fatalf("New() error %v", err)
			}
			go func() {
				for {
					select {
					case <-l.Channel():
					case <-ec:
						return
					}
				}
			}()
			for k, v := range tt.addCalls {
				if err := l.Add(v.first); (err != nil) != tt.wantErrs[k] {
					t.Errorf("Provider.Add() error = %v, wantErr %v", err, tt.wantErrs[k])
				}
			}
			if !reflect.DeepEqual(r.addCalls, tt.addCalls) {
				t.Errorf("Reader.Add() got %v, want %v", r.addCalls, tt.addCalls)
			}

			var wwc []ArgsStr
			for i := range tt.addCalls {
				if tt.readerAddRets[i].first == nil {
					wwc = append(wwc, tt.addCalls[i])
				}
			}
			if !reflect.DeepEqual(w.addCalls, wwc) {
				t.Errorf("Watcher.Add() got %v, want %v", w.addCalls, wwc)
			}
			ec <- struct{}{}
		})
	}
}

func TestProvider(t *testing.T) {
	tests := []struct {
		name string
		vals []struct {
			readerVal  ReaderValue
			readerErr  error
			watcherErr error
		}
	}{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var rr []ArgsReaderValueErr
			for _, v := range tt.vals {
				rr = append(rr, ArgsReaderValueErr{
					first:  v.readerVal,
					second: v.readerErr,
				})
			}
			rf := NewReaderDoubleFactoryFunc([]ArgsErr{{nil}}, []ArgsErr{{nil}}, rr)
			wf := NewWatcherDoubleFactoryFunc([]ArgsErr{{nil}}, []ArgsErr{{}})
			ri, _ := rf()
			wi, _ := wf()
			r := ri.(*ReaderDouble)
			w := wi.(*WatcherDouble)

			ec := make(chan struct{})
			l, err := New(rf, wf)
			if err != nil {
				t.Fatalf("New() error %v", err)
			}

			var gotVals []ReaderValueWithError
			go func() {
				for {
					select {
					case v := <-l.Channel():
						gotVals = append(gotVals, v)
					case <-ec:
						return
					}
				}
			}()
			if err := l.Add("foo"); err != nil {
				t.Fatalf("Provider.Add() error %v", err)
			}
			for _, v := range tt.vals {
				w.ch <- WatcherMsg{
					Path: "foo",
					Err:  v.watcherErr,
				}
			}

			ec <- struct{}{}
		})
	}
}
