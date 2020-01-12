package rlistener

import (
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/pkg/errors"
)

type Listener struct {
	w   Watcher
	ch  chan ListenerMsg
	mtx sync.Mutex
	ld  map[string][]string
}

type ListenerMsg struct {
	Path string
	Op   Operation
	Err  error
}

type WatcherFactoryFunc func() (Watcher, error)

func (l *Listener) AddDir(path string) error {
	i, err := os.Stat(path)
	if err != nil {
		return err
	}
	if !i.IsDir() {
		return errors.New("the path did not point to a directory")
	}
	path, err = filepath.Abs(path)
	if err != nil {
		return err
	}
	return l.w.AddDir(path)
}

func (l *Listener) Channel() <-chan ListenerMsg {
	return l.ch
}

func (l *Listener) Close() error {
	return l.w.Close()
}

func pathDiff(src, dst []string) (rem []string, add []string) {
	// can be vastly improved by using binary search
	// because the paths are in lexical order
	for i := 0; i < len(src); i++ {
		path := src[i]
		for j := 0; ; j++ {
			if path == dst[j] {
				break
			}
			if j == len(dst)-1 {
				rem = append(rem, path)
				break
			}
		}
	}
	for i := 0; i < len(dst); i++ {
		path := dst[i]
		for j := 0; ; j++ {
			if path == src[j] {
				break
			}
			if j == len(src)-1 {
				add = append(add, path)
				break
			}
		}
	}
	return
}

func (l *Listener) applyDiff(add []string, rem []string) error {
	for _, v := range rem {
		if err := l.w.RemDir(v); err != nil {
			return err
		}
	}
	for _, v := range add {
		if err := l.w.AddDir(v); err != nil {
			return err
		}
	}
	return nil
}

func (l *Listener) recursiveDiscover(path string) error {
	var dst []string
	err := filepath.Walk(path, func(fpath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			return nil
		}
		r, err := filepath.Rel(path, fpath)
		if err != nil {
			return err
		}
		dst = append(dst, r)
		return nil
	})
	if err != nil {
		return err
	}
	l.mtx.Lock()
	defer l.mtx.Unlock()
	src, _ := l.ld[path]
	if err := l.applyDiff(pathDiff(src, dst)); err != nil {
		return err
	}
	l.ld[path] = dst
	return nil
}

func (l *Listener) handleMsg(w ListenerMsg) {
	for k := range l.ld {
		if strings.HasPrefix(w.Path, k) {
			l.recursiveDiscover(k)
		}
	}
}

func (l *Listener) run() {
	c := l.w.Channel()
	for {
		w, ok := <-c
		if !ok {
			break
		}
		if w.Err != nil {
			l.ch <- ListenerMsg{Err: w.Err}
			continue
		}
		ap, err := filepath.Abs(w.Path)
		if err != nil {
			l.ch <- ListenerMsg{Err: w.Err}
			continue
		}
		m := ListenerMsg{
			Path: ap,
			Op:   w.Op,
			Err:  w.Err,
		}
		l.handleMsg(m)
		l.ch <- m
	}
	close(l.ch)
}

func New(watcherFactory WatcherFactoryFunc) (*Listener, error) {
	if watcherFactory == nil {
		return nil, errors.New("watcherFactory cannot be nil")
	}

	w, err := watcherFactory()
	if err != nil {
		return nil, errors.Wrap(err, "failed to create watcher")
	}
	l := &Listener{
		w:  w,
		ch: make(chan ListenerMsg),
	}
	go l.run()
	return l, nil
}
