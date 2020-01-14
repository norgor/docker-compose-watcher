package provider

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
			readerFactory:    NewReaderDoubleFactoryFunc([]ArgsErr{}, []ArgsErr{{nil}}, []ArgsReaderValueErr{}),
			watcherFactory:   NewWatcherDoubleFactoryFunc([]ArgsErr{}, []ArgsErr{{nil}}),
			wantChannelClose: true,
			wantErr:          false,
		},
		{
			name:             "closes channel and returns error when reader error occurs",
			readerFactory:    NewReaderDoubleFactoryFunc([]ArgsErr{}, []ArgsErr{{fmt.Errorf("foo")}}, []ArgsReaderValueErr{}),
			watcherFactory:   NewWatcherDoubleFactoryFunc([]ArgsErr{}, []ArgsErr{{nil}}),
			wantChannelClose: true,
			wantErr:          true,
		},
		{
			name:             "closes channel and returns error when watcher error occurs",
			readerFactory:    NewReaderDoubleFactoryFunc([]ArgsErr{}, []ArgsErr{{nil}}, []ArgsReaderValueErr{}),
			watcherFactory:   NewWatcherDoubleFactoryFunc([]ArgsErr{}, []ArgsErr{{fmt.Errorf("foo")}}),
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
				{nil},
				{nil},
			},
			readerAddRets: []ArgsErr{
				{nil},
				{nil},
			},
			addCalls: []ArgsStr{
				{"foo"},
				{"bar"},
			},
			wantErrs: []bool{false, false},
		},
		{
			name:           "returns err when Watcher.Add() errors",
			watcherAddRets: []ArgsErr{{fmt.Errorf("foo")}},
			readerAddRets:  []ArgsErr{{nil}},
			addCalls:       []ArgsStr{{"foo"}},
			wantErrs:       []bool{true},
		},
		{
			name:           "returns err when Reader.Add() errors",
			watcherAddRets: []ArgsErr{{nil}},
			readerAddRets:  []ArgsErr{{fmt.Errorf("foo")}},
			addCalls:       []ArgsStr{{"foo"}},
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

			l, err := New(rf, wf)
			if err != nil {
				t.Fatalf("New() error %v", err)
			}

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
		})
	}
}

func TestProvider(t *testing.T) {
	type val struct {
		readerVal  ReaderValue
		readerErr  error
		watcherErr error
	}
	tests := []struct {
		name string
		vals []val
	}{
		{
			name: "test happy",
			vals: []val{
				{readerVal: "foo"},
				{readerVal: "bar"},
				{readerVal: "baz"},
			},
		},
		{
			name: "reader error",
			vals: []val{
				{readerErr: fmt.Errorf("foo")},
				{readerErr: fmt.Errorf("bar")},
				{readerErr: fmt.Errorf("baz")},
			},
		},
		{
			name: "watcher error",
			vals: []val{
				{watcherErr: fmt.Errorf("foo")},
				{watcherErr: fmt.Errorf("bar")},
				{watcherErr: fmt.Errorf("baz")},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			//one empty needed because Add() will cause a sync
			wantVals := []ReaderValueWithError{{}}
			rr := []ArgsReaderValueErr{{}}
			for _, v := range tt.vals {
				rr = append(rr, ArgsReaderValueErr{
					first:  v.readerVal,
					second: v.readerErr,
				})
				var x ReaderValueWithError
				if v.watcherErr != nil {
					x = ReaderValueWithError{Error: v.watcherErr}
				} else {
					x = ReaderValueWithError{
						Value: v.readerVal,
						Error: v.readerErr,
					}
				}
				wantVals = append(wantVals, x)
			}
			rf := NewReaderDoubleFactoryFunc([]ArgsErr{{nil}}, []ArgsErr{{nil}}, rr)
			wf := NewWatcherDoubleFactoryFunc([]ArgsErr{{nil}}, []ArgsErr{{}})
			wi, _ := wf()
			w := wi.(*WatcherDouble)

			l, err := New(rf, wf)
			if err != nil {
				t.Fatalf("New() error %v", err)
			}
			ws := make(chan struct{})
			cc := make(chan struct{})
			var gotVals []ReaderValueWithError
			go func() {
				for {
					select {
					case v, ok := <-l.Channel():
						if !ok {
							cc <- struct{}{}
							return
						}
						gotVals = append(gotVals, v)
						select {
						case ws <- struct{}{}:
						default:
						}
					}
				}
			}()
			if err := l.Add("foo"); err != nil {
				t.Fatalf("Provider.Add() error %v", err)
			}
			<-ws
			for _, v := range tt.vals {
				w.ch <- WatcherMsg{
					Path: "foo",
					Err:  v.watcherErr,
				}

			}
			close(w.ch)
			<-cc

			if !reflect.DeepEqual(gotVals, wantVals) {
				t.Errorf("Provider.Channel() Messages got %v, want %v", gotVals, wantVals)
			}
		})
	}
}
