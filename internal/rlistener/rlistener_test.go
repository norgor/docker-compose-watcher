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

func TestListenerWithWatchers(t *testing.T) {
	f := map[string]rlistener.WatcherFactoryFunc{
		"fsnotify": fsnotify.New,
	}
	for k, v := range f {
		t.Run(k, func(tt *testing.T) {
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
			if err != nil {
				tt.Errorf("New() error %v", err)
			}
			path, err := ioutil.TempDir("", "rlistener_test")
			if err != nil {
				tt.Fatalf("failed to create temp dir, error %v", err)
			}
			defer func() {
				if err := os.RemoveAll(path); err != nil {
					tt.Fatalf("failed to remove temp dir, error %v", err)
				}
			}()
			if err := l.AddDir(path); err != nil {
				tt.Errorf("Listener.AddDir() error %v", err)
			}

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
			if err := os.MkdirAll(ap, 0700); err != nil {
				tt.Fatalf("failed to create test directories, error %v", err)
			}
			for k := range want {
				want[k].Path = filepath.Join(path, want[k].Path)
			}
			<-n
			ioutil.WriteFile(filepath.Join(ap, "file"), []byte("foo"), 0700)
			if err := os.RemoveAll(filepath.Join(path, "/0/1")); err != nil {
				tt.Errorf("failed to remove subdirs, error %v", err)
			}

			if err := l.Close(); err != nil {
				tt.Errorf("Listener.Close() error %v", err)
			}
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
