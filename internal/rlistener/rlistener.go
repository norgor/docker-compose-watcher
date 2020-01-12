package rlistener

import (
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/pkg/errors"
)

// Listener recursively listens for changes within added directories.
type Listener struct {
	w   Watcher
	ch  chan ListenerMsg
	mtx sync.Mutex
	ld  map[string][]string
}

// ListenerMsg is a message from the listener.
type ListenerMsg struct {
	Path      string
	Operation Operation
	Error     error
}

// WatcherFactoryFunc is a function for creting watchers for the listener.
type WatcherFactoryFunc func() (Watcher, error)

// AddDir adds a directory to listen on.
func (l *Listener) AddDir(path string) error {
	i, err := os.Stat(path)
	if err != nil {
		return errors.Wrapf(err, "stat failed on file %s", path)
	}
	if !i.IsDir() {
		return errors.New("the path did not point to a directory")
	}
	path, err = filepath.Abs(path)
	if err != nil {
		return errors.Wrapf(err, "failed to get absolute path of %s", path)
	}
	if err := l.recursiveDiscover(path); err != nil {
		return err
	}
	return l.w.AddDir(path)
}

// Channel returns the listener's channel.
func (l *Listener) Channel() <-chan ListenerMsg {
	return l.ch
}

// Close closes the listener.
func (l *Listener) Close() error {
	return l.w.Close()
}

func pathDiff(src, dst []string) (add []string, rem []string) {
	// can be vastly improved by using binary search
	// because the paths are in lexical order
	for i := 0; i < len(src); i++ {
		path := src[i]
		for j := 0; ; j++ {
			if j == len(dst) {
				rem = append(rem, path)
				break
			}
			if path == dst[j] {
				break
			}
		}
	}
	for i := 0; i < len(dst); i++ {
		path := dst[i]
		for j := 0; ; j++ {
			if j == len(src) {
				add = append(add, path)
				break
			}
			if path == src[j] {
				break
			}
		}
	}
	return
}

func (l *Listener) applyDiff(path string, add []string, rem []string) error {
	for _, v := range rem {
		if err := l.w.RemDir(filepath.Join(path, v)); err != nil {
			return errors.Wrap(err, "failed to remove dir")
		}
	}
	for _, v := range add {
		if err := l.w.AddDir(filepath.Join(path, v)); err != nil {
			return errors.Wrap(err, "failed to add dir")
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
			return errors.Wrap(err, "failed to get relative path")
		}
		dst = append(dst, r)
		return nil
	})
	if err != nil {
		return errors.Wrap(err, "recursive walk discovery failed")
	}
	l.mtx.Lock()
	defer l.mtx.Unlock()
	src, _ := l.ld[path]
	add, rem := pathDiff(src, dst)
	if err := l.applyDiff(path, add, rem); err != nil {
		return errors.Wrap(err, "failed to apply diff")
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
			l.ch <- ListenerMsg{Error: w.Err}
			continue
		}
		ap, err := filepath.Abs(w.Path)
		if err != nil {
			l.ch <- ListenerMsg{Error: w.Err}
			continue
		}
		m := ListenerMsg{
			Path:      ap,
			Operation: w.Op,
			Error:     w.Err,
		}
		l.handleMsg(m)
		l.ch <- m
	}
	close(l.ch)
}

// New creates a new rlistener
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
		ld: make(map[string][]string),
	}
	go l.run()
	return l, nil
}
