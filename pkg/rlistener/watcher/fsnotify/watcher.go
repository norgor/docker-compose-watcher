package fsnotify

import (
	"docker-compose-watcher/pkg/rlistener"
	"github.com/fsnotify/fsnotify"
)

// Watcher watcher for file changes.
type Watcher struct {
	*fsnotify.Watcher
	ch chan rlistener.WatcherMsg
}

// Chan returns the watcher channel.
func (w *Watcher) Channel() <-chan rlistener.WatcherMsg {
	return w.ch
}

func (w *Watcher) AddDir(path string) error {
	return w.Add(path)
}

func (w *Watcher) RemDir(path string) error {
	return w.Remove(path)
}

// New creates a new fsnotify watcher.
func New() (rlistener.Watcher, error) {
	w, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}
	c := make(chan rlistener.WatcherMsg)
	go func() {
	loop:
		for {
			select {
			case err, ok := <-w.Errors:
				if !ok {
					break loop
				}
				c <- rlistener.WatcherMsg{
					Path: "",
					Err:  err,
				}
			case e, ok := <-w.Events:
				if !ok {
					break loop
				}
				c <- rlistener.WatcherMsg{
					Path: e.Name,
					Err:  nil,
				}
			}
		}
		close(c)
	}()
	return &Watcher{
		w,
		c,
	}, nil
}
