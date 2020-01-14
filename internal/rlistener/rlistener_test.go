package rlistener_test

import (
	"docker-compose-watcher/internal/rlistener"
	"docker-compose-watcher/internal/rlistener/watcher/fsnotify"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

type TExtended testing.T

func (t *TExtended) errorIfErr(err error, msg string) {
	t.Helper()
	if err != nil {
		t.Errorf("%s error %v", msg, err)
	}
}

func (t *TExtended) fatalIfErr(err error, msg string) {
	t.Helper()
	if err != nil {
		t.Fatalf("%s error %v", msg, err)
	}
}

func TestListenerWithWatchers(t *testing.T) {
	f := map[string]rlistener.WatcherFactoryFunc{
		"fsnotify": fsnotify.New,
	}
	for k, v := range f {
		t.Run(k, func(tts *testing.T) {
			tt := (*TExtended)(tts)
			want := []rlistener.ListenerMsg{
				{
					Path:      "./0",
					Operation: rlistener.Create,
					Error:     nil,
				},
				{
					Path:      "./0/1/2/3/4/5/6/7/8/9/10/11/12/13/14/15/16/file",
					Operation: rlistener.Create,
					Error:     nil,
				},
				{
					Path:      "./0/1/2/3/4/5/6/7/8/9/10/11/12/13/14/15/16/file",
					Operation: rlistener.Write,
					Error:     nil,
				},
			}
			l, err := rlistener.New(v)
			tt.errorIfErr(err, "New()")

			path, err := ioutil.TempDir("", "rlistener_test")
			tt.fatalIfErr(err, "failed to create temp dir")
			defer func() {
				tt.fatalIfErr(os.RemoveAll(path), "failed to remove temp dir")
			}()
			tt.errorIfErr(l.AddDir(path), "Listener.AddDir()")

			var got []rlistener.ListenerMsg
			n := make(chan struct{})
			done := make(chan struct{})
			go func() {
				c := l.Channel()
				for {
					m, ok := <-c
					if !ok {
						break
					}
					got = append(got, m)
					select {
					case n <- struct{}{}:
					default:
					}
				}
				done <- struct{}{}
			}()
			ap := filepath.Join(path, "/0/1/2/3/4/5/6/7/8/9/10/11/12/13/14/15/16")
			tt.errorIfErr(os.MkdirAll(ap, 0700), "failed to create test directories")
			for k := range want {
				want[k].Path = filepath.Join(path, want[k].Path)
			}
			<-n
			err = ioutil.WriteFile(filepath.Join(ap, "file"), []byte("foo"), 0700)
			tt.errorIfErr(err, "failed to write to file")
			tt.errorIfErr(
				os.RemoveAll(filepath.Join(path, "/0/1")),
				"failed to remove subdirs",
			)
			tt.errorIfErr(l.Close(), "Listener.Close()")
			<-done
			for _, v := range want {
				for i := 0; ; i++ {
					if i == len(got) {
						tt.Errorf("Listener.Channel() %v was not sent", v)
						break
					}
					if reflect.DeepEqual(v, got[i]) {
						break
					}
				}
			}
		})
	}
}
